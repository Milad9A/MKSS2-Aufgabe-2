package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin mode for production
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add middleware to detect HTTPS from headers (for proxy/load balancer scenarios)
	router.Use(func(c *gin.Context) {
		// Check for common HTTPS detection headers
		if c.GetHeader("X-Forwarded-Proto") == "https" ||
			c.GetHeader("X-Forwarded-SSL") == "on" ||
			c.GetHeader("X-URL-Scheme") == "https" {
			c.Set("scheme", "https")
		} else if c.Request.TLS != nil {
			c.Set("scheme", "https")
		} else {
			c.Set("scheme", "http")
		}
		c.Next()
	})

	// Configure CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                                       // Allow all origins
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},   // Allowed methods
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // Allowed headers
		ExposeHeaders:    []string{"Content-Length"},                          // Exposed headers
		AllowCredentials: true,                                                // Allow cookies
		MaxAge:           12 * time.Hour,                                      // Preflight request cache duration
	}))

	// Add enhanced health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
			"service":   "robot-api",
		})
	})

	// Add a simple root endpoint for basic connectivity test
	router.GET("/", func(c *gin.Context) {
		scheme := c.GetString("scheme")
		if scheme == "" {
			scheme = "http"
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Robot API Server is running",
			"version":       "1.0.0",
			"scheme":        scheme,
			"https_enabled": scheme == "https",
			"endpoints": []string{
				"/health",
				"/robot/{id}/status",
				"/robot/{id}/move",
				"/robot/{id}/pickup/{itemId}",
				"/robot/{id}/putdown/{itemId}",
				"/robot/{id}/state",
				"/robot/{id}/actions",
				"/robot/{id}/attack/{targetId}",
			},
		})
	})

	storage := NewRobotStorage()
	storage.Initialize()
	handler := NewRobotHandler(storage)

	// Add items endpoint to check available items
	router.GET("/items", func(c *gin.Context) {
		items := storage.GetAvailableItems()
		c.JSON(http.StatusOK, gin.H{
			"available_items": items,
			"total_count":     len(items),
		})
	})

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

	// Get port from environment variable, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting robot API server on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give the server 5 seconds to finish any ongoing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
