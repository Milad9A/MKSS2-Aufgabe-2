package main

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RobotHandler handles robot-related requests
type RobotHandler struct {
	storage *RobotStorage
}

// NewRobotHandler creates a new handler with the given storage
func NewRobotHandler(storage *RobotStorage) *RobotHandler {
	return &RobotHandler{storage: storage}
}

// GetStatus returns the current status of a robot
func (h *RobotHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	robot, err := h.storage.GetRobot(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Robot not found"})
		return
	}

	// Create HATEOAS links
	baseURL := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	links := []Link{
		{
			Rel:  "self",
			Href: fmt.Sprintf("%s://%s/robot/%s/status", scheme, baseURL, id),
		},
		{
			Rel:  "actions",
			Href: fmt.Sprintf("%s://%s/robot/%s/actions?page=1&size=5", scheme, baseURL, id),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        robot.ID,
		"position":  robot.Position,
		"energy":    robot.Energy,
		"inventory": robot.Inventory,
		"links":     links,
	})
}

// MoveRobot moves a robot in the specified direction
func (h *RobotHandler) MoveRobot(c *gin.Context) {
	id := c.Param("id")
	robot, err := h.storage.GetRobot(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Robot not found"})
		return
	}

	var moveReq MoveRequest
	if err := c.ShouldBindJSON(&moveReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Apply movement based on direction
	switch moveReq.Direction {
	case "up":
		robot.Position.Y++
	case "down":
		robot.Position.Y--
	case "left":
		robot.Position.X--
	case "right":
		robot.Position.X++
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid direction"})
		return
	}

	// Add action to history
	h.storage.AddAction(id, "move", fmt.Sprintf("Moved %s", moveReq.Direction))
	h.storage.SaveRobot(robot)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Robot moved successfully",
		"position": robot.Position,
	})
}

// PickupItem allows a robot to pick up an item
func (h *RobotHandler) PickupItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")

	robot, err := h.storage.GetRobot(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Robot not found"})
		return
	}

	if !h.storage.ItemExists(itemID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Add item to inventory
	robot.Inventory = append(robot.Inventory, itemID)
	h.storage.RemoveItem(itemID) // Remove from world
	h.storage.SaveRobot(robot)
	h.storage.AddAction(id, "pickup", fmt.Sprintf("Picked up item %s", itemID))

	c.JSON(http.StatusOK, gin.H{
		"message":   "Item picked up successfully",
		"inventory": robot.Inventory,
	})
}

// PutdownItem allows a robot to put down an item
func (h *RobotHandler) PutdownItem(c *gin.Context) {
	id := c.Param("id")
	itemID := c.Param("itemId")

	robot, err := h.storage.GetRobot(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Robot not found"})
		return
	}

	// Check if robot has the item
	hasItem := false
	var newInventory []string
	for _, item := range robot.Inventory {
		if item == itemID {
			hasItem = true
		} else {
			newInventory = append(newInventory, item)
		}
	}

	if !hasItem {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Robot does not have this item"})
		return
	}

	// Update robot and world
	robot.Inventory = newInventory
	h.storage.AddItem(itemID)
	h.storage.SaveRobot(robot)
	h.storage.AddAction(id, "putdown", fmt.Sprintf("Put down item %s", itemID))

	c.JSON(http.StatusOK, gin.H{
		"message":   "Item put down successfully",
		"inventory": robot.Inventory,
	})
}

// UpdateState updates a robot's state
func (h *RobotHandler) UpdateState(c *gin.Context) {
	id := c.Param("id")
	robot, err := h.storage.GetRobot(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Robot not found"})
		return
	}

	var stateReq StateUpdateRequest
	if err := c.ShouldBindJSON(&stateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Update energy if provided
	if stateReq.Energy != nil {
		robot.Energy = *stateReq.Energy
		h.storage.AddAction(id, "update", fmt.Sprintf("Updated energy to %d", *stateReq.Energy))
	}

	// Update position if provided
	if stateReq.Position != nil {
		robot.Position = *stateReq.Position
		h.storage.AddAction(id, "update", fmt.Sprintf("Updated position to (%d,%d)",
			stateReq.Position.X, stateReq.Position.Y))
	}

	h.storage.SaveRobot(robot)

	c.JSON(http.StatusOK, gin.H{
		"message": "Robot state updated successfully",
		"robot":   robot,
	})
}

// GetActions returns all actions performed by a robot with pagination
func (h *RobotHandler) GetActions(c *gin.Context) {
	id := c.Param("id")
	robot, err := h.storage.GetRobot(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Robot not found"})
		return
	}

	// Get pagination parameters
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "5")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 5
	}

	// Calculate pagination
	totalElements := len(robot.Actions)
	totalPages := int(math.Ceil(float64(totalElements) / float64(size)))

	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	startIndex := (page - 1) * size
	endIndex := startIndex + size
	if endIndex > totalElements {
		endIndex = totalElements
	}

	// Create paginated actions slice
	var paginatedActions []ActionWithLinks
	for i := startIndex; i < endIndex; i++ {
		action := robot.Actions[i]
		actionWithLinks := ActionWithLinks{
			Action: action,
			Links: []Link{
				{
					Rel:  "self",
					Href: fmt.Sprintf("http://%s/robot/%s/actions/%d", c.Request.Host, id, i+1),
				},
			},
		}
		paginatedActions = append(paginatedActions, actionWithLinks)
	}

	// Create page info
	pageInfo := PageInfo{
		Number:        page,
		Size:          size,
		TotalElements: totalElements,
		TotalPages:    totalPages,
		HasNext:       page < totalPages,
		HasPrevious:   page > 1,
	}

	// Create navigation links
	var links []Link
	if pageInfo.HasNext {
		links = append(links, Link{
			Rel:  "next",
			Href: fmt.Sprintf("/robot/%s/actions?page=%d&size=%d", id, page+1, size),
		})
	}

	if pageInfo.HasPrevious {
		links = append(links, Link{
			Rel:  "previous",
			Href: fmt.Sprintf("/robot/%s/actions?page=%d&size=%d", id, page-1, size),
		})
	}

	response := PaginatedActions{
		Page:    pageInfo,
		Actions: paginatedActions,
		Links:   links,
	}

	c.JSON(http.StatusOK, response)
}

// AttackRobot handles one robot attacking another
func (h *RobotHandler) AttackRobot(c *gin.Context) {
	id := c.Param("id")
	targetID := c.Param("targetId")

	// Get attacker
	attacker, err := h.storage.GetRobot(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Attacker robot not found"})
		return
	}

	// Get target
	target, err := h.storage.GetRobot(targetID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Target robot not found"})
		return
	}

	// Cost for attacker (5% energy)
	energyReduction := attacker.Energy * 5 / 100
	attacker.Energy -= energyReduction

	// Generate random damage to target (10-20% energy)
	damageFactor := 15 // 15% damage for simplicity
	damage := target.Energy * damageFactor / 100
	target.Energy -= damage

	// Ensure energy doesn't go below 0
	if target.Energy < 0 {
		target.Energy = 0
	}

	// Save changes
	h.storage.AddAction(id, "attack", fmt.Sprintf("Attacked robot %s", targetID))
	h.storage.AddAction(targetID, "damaged", fmt.Sprintf("Damaged by robot %s", id))
	h.storage.SaveRobot(attacker)
	h.storage.SaveRobot(target)

	c.JSON(http.StatusOK, gin.H{
		"message":         "Attack successful",
		"attacker_energy": attacker.Energy,
		"target_energy":   target.Energy,
		"damage_dealt":    damage,
	})
}
