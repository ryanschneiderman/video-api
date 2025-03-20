package main

import (
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ryanschneiderman/video-api/internal/app"
	"github.com/ryanschneiderman/video-api/internal/handlers"
	"github.com/ryanschneiderman/video-api/internal/metrics"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env variables")
	}
	reg := prometheus.NewRegistry()
	apiMetrics := metrics.NewAPIMetrics()
	apiMetrics.Register(reg)


	ctx := context.TODO()
	a, err := app.InitializeApp(ctx)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	router := setupRouter(a, reg, apiMetrics)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("Starting server on port:", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRouter(a *app.App, reg *prometheus.Registry, apiMetrics *metrics.APIMetrics) *gin.Engine {
	router := gin.Default()
	router.Use(metrics.PrometheusMiddleware(apiMetrics))

	videoHandler := handlers.NewVideoHandler(a)
	router.POST("/videos", videoHandler.UploadVideo)
	router.GET("/videos/:id", videoHandler.GetVideo)
	// Serve metrics from the provided custom registry.
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})))
	return router
}
