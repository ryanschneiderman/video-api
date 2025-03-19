package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/ryanschneiderman/video-api/internal/app"
	"github.com/ryanschneiderman/video-api/internal/handlers"
	"github.com/ryanschneiderman/video-api/internal/metrics"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env variables")
	}

	metrics.RegisterMetrics()

	// Initialize the shared App with AWS clients and configuration
	ctx := context.TODO()
	app, err := app.InitializeApp(ctx)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Set up Gin router
	router := gin.Default()
	router.Use(metrics.PrometheusMiddleware())

	// Create a VideoHandler that uses the shared App dependencies
	videoHandler := handlers.NewVideoHandler(app)

	// Define routes
	router.POST("/videos", videoHandler.UploadVideo)
	router.GET("/videos/:id", videoHandler.GetVideo)

	router.GET("/metrics", gin.WrapH(metrics.MetricsHandler()))

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Starting server on port:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
