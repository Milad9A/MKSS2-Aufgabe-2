package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

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

	log.Println("Starting robot API server on port 8080...")
	router.Run(":8080")
}
