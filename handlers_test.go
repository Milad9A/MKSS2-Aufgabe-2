package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() (*gin.Engine, *RobotStorage) {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	storage := NewRobotStorage()
	storage.Initialize()
	handler := NewRobotHandler(storage)

	api := router.Group("/robot")
	{
		api.GET("/:id/status", handler.GetStatus)
		api.POST("/:id/move", handler.MoveRobot)
		api.POST("/:id/pickup/:itemId", handler.PickupItem)
		api.POST("/:id/putdown/:itemId", handler.PutdownItem)
		api.PATCH("/:id/state", handler.UpdateState)
		api.GET("/:id/actions", handler.GetActions)
		api.POST("/:id/attack/:targetId", handler.AttackRobot)
	}

	return router, storage
}

func TestGetStatus(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/robot/robot1/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "robot1", response["id"])
	assert.NotNil(t, response["position"])
	assert.NotNil(t, response["energy"])
	assert.NotNil(t, response["inventory"])
	assert.NotNil(t, response["links"])

	links, ok := response["links"].([]interface{})
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(links), 1)
}

func TestMoveRobot(t *testing.T) {
	router, storage := setupTestRouter()

	robot, _ := storage.GetRobot("robot1")
	initialX := robot.Position.X
	initialY := robot.Position.Y

	moveBody := `{"direction": "up"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/robot/robot1/move", bytes.NewBufferString(moveBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "successfully")

	robot, _ = storage.GetRobot("robot1")
	assert.Equal(t, initialX, robot.Position.X)
	assert.Equal(t, initialY+1, robot.Position.Y)
}

func TestPickupItem(t *testing.T) {
	router, storage := setupTestRouter()

	storage.AddItem("item1")

	robot, _ := storage.GetRobot("robot1")
	initialInventorySize := len(robot.Inventory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/robot/robot1/pickup/item1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "successfully")

	robot, _ = storage.GetRobot("robot1")
	assert.Equal(t, initialInventorySize+1, len(robot.Inventory))
	assert.Contains(t, robot.Inventory, "item1")

	assert.False(t, storage.ItemExists("item1"))
}

func TestPutdownItem(t *testing.T) {
	router, storage := setupTestRouter()

	robot, _ := storage.GetRobot("robot1")
	storage.RemoveItem("item2")
	robot.Inventory = append(robot.Inventory, "item2")
	storage.SaveRobot(robot)

	initialInventorySize := len(robot.Inventory)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/robot/robot1/putdown/item2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "successfully")

	robot, _ = storage.GetRobot("robot1")
	assert.Equal(t, initialInventorySize-1, len(robot.Inventory))
	assert.NotContains(t, robot.Inventory, "item2")

	assert.True(t, storage.ItemExists("item2"))
}

func TestUpdateState(t *testing.T) {
	router, storage := setupTestRouter()

	updateBody := `{"energy": 75, "position": {"x": 5, "y": 8}}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/robot/robot1/state", bytes.NewBufferString(updateBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "successfully")

	robot, _ := storage.GetRobot("robot1")
	assert.Equal(t, 75, robot.Energy)
	assert.Equal(t, 5, robot.Position.X)
	assert.Equal(t, 8, robot.Position.Y)
}

func TestGetActions(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/robot/robot1/actions?size=2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response PaginatedActions
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 1, response.Page.Number)
	assert.Equal(t, 2, response.Page.Size)
	assert.True(t, response.Page.HasNext)
	assert.False(t, response.Page.HasPrevious)

	assert.Len(t, response.Actions, 2)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/robot/robot1/actions?page=2&size=2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 2, response.Page.Number)
	assert.True(t, response.Page.HasNext)
	assert.True(t, response.Page.HasPrevious)
	assert.Len(t, response.Actions, 2)
}

func TestAttackRobot(t *testing.T) {
	router, storage := setupTestRouter()

	attacker, _ := storage.GetRobot("robot1")
	target, _ := storage.GetRobot("robot2")
	attackerEnergy := attacker.Energy
	targetEnergy := target.Energy

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/robot/robot1/attack/robot2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Contains(t, response["message"], "successful")
	assert.Contains(t, response, "damage_dealt")

	attacker, _ = storage.GetRobot("robot1")
	target, _ = storage.GetRobot("robot2")

	assert.Less(t, attacker.Energy, attackerEnergy)

	assert.Less(t, target.Energy, targetEnergy)
}

func TestRobotNotFound(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/robot/nonexistent/status", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "not found")
}
