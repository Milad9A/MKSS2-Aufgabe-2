package main

import (
	"errors"
	"sync"
	"time"
)

// RobotStorage provides in-memory storage for robots
type RobotStorage struct {
	robots map[string]*Robot
	items  map[string]bool
	mutex  sync.RWMutex
}

// NewRobotStorage creates a new instance of RobotStorage
func NewRobotStorage() *RobotStorage {
	return &RobotStorage{
		robots: make(map[string]*Robot),
		items:  make(map[string]bool),
	}
}

// GetRobot retrieves a robot by ID
func (s *RobotStorage) GetRobot(id string) (*Robot, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	robot, exists := s.robots[id]
	if !exists {
		return nil, errors.New("robot not found")
	}
	return robot, nil
}

// SaveRobot saves a robot to storage
func (s *RobotStorage) SaveRobot(robot *Robot) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.robots[robot.ID] = robot
}

// AddAction adds an action to a robot's history
func (s *RobotStorage) AddAction(robotID, actionType, details string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	robot, exists := s.robots[robotID]
	if !exists {
		return errors.New("robot not found")
	}

	action := Action{
		Type:      actionType,
		Timestamp: time.Now(),
		Details:   details,
	}

	robot.Actions = append(robot.Actions, action)
	return nil
}

// ItemExists checks if an item exists in the world
func (s *RobotStorage) ItemExists(itemID string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.items[itemID]
}

// AddItem adds an item to the world
func (s *RobotStorage) AddItem(itemID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.items[itemID] = true
}

// RemoveItem removes an item from the world
func (s *RobotStorage) RemoveItem(itemID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.items, itemID)
}

// Initialize storage with some example data
func (s *RobotStorage) Initialize() {
	// Create some example robots
	robot1 := &Robot{
		ID:        "robot1",
		Position:  Position{X: 0, Y: 0},
		Direction: "north",
		Energy:    100,
		Inventory: []string{},
		Actions: []Action{
			{
				Type:      "create",
				Timestamp: time.Now().Add(-24 * time.Hour),
				Details:   "Robot was created",
			},
			{
				Type:      "move",
				Timestamp: time.Now().Add(-12 * time.Hour),
				Details:   "Moved north",
			},
			{
				Type:      "pickup",
				Timestamp: time.Now().Add(-6 * time.Hour),
				Details:   "Picked up item1",
			},
			{
				Type:      "putdown",
				Timestamp: time.Now().Add(-3 * time.Hour),
				Details:   "Put down item1",
			},
			{
				Type:      "update",
				Timestamp: time.Now().Add(-1 * time.Hour),
				Details:   "Updated energy to 100",
			},
			{
				Type:      "move",
				Timestamp: time.Now().Add(-30 * time.Minute),
				Details:   "Moved east",
			},
			{
				Type:      "attack",
				Timestamp: time.Now().Add(-15 * time.Minute),
				Details:   "Attacked robot2",
			},
		},
	}

	robot2 := &Robot{
		ID:        "robot2",
		Position:  Position{X: 10, Y: 10},
		Direction: "south",
		Energy:    100,
		Inventory: []string{},
		Actions: []Action{
			{
				Type:      "create",
				Timestamp: time.Now().Add(-24 * time.Hour),
				Details:   "Robot was created",
			},
			{
				Type:      "move",
				Timestamp: time.Now().Add(-10 * time.Hour),
				Details:   "Moved south",
			},
			{
				Type:      "damaged",
				Timestamp: time.Now().Add(-15 * time.Minute),
				Details:   "Damaged by robot1",
			},
		},
	}

	s.items["item1"] = true
	s.items["item2"] = true
	s.items["item3"] = true

	s.robots["robot1"] = robot1
	s.robots["robot2"] = robot2
}
