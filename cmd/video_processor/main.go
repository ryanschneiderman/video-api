package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/ryanschneiderman/video-api/internal/app"
	"github.com/ryanschneiderman/video-api/internal/metrics"
	"github.com/ryanschneiderman/video-api/internal/worker"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Register Prometheus metrics
	metrics.RegisterMetrics()

	// Start a metrics server in a separate goroutine so Prometheus can scrape metrics.
	go func() {
		http.Handle("/metrics", metrics.MetricsHandler())
		log.Println("Starting metrics server on :9090")
		log.Fatal(http.ListenAndServe(":9090", nil))
	}()

	// Initialize the shared app (AWS clients, configuration, etc.)
	ctx := context.Background()
	myApp, err := app.InitializeApp(ctx)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Create a new worker processor with the shared dependencies
	processor := worker.NewProcessor(myApp)

	// Start polling SQS continuously, with instrumentation around the processing.
	log.Println("Starting SQS worker...")
	for {
		start := time.Now()

		// Process messages from SQS; update custom metrics if needed.
		if err := processor.ProcessMessages(ctx); err != nil {
			log.Printf("Error processing messages: %v", err)
			// Optionally increment an error counter metric.
			metrics.WorkerErrors.Inc()
		}

		duration := time.Since(start).Seconds()
		// Record processing duration for this iteration.
		metrics.WorkerProcessingDuration.Observe(duration)

		// Sleep briefly to avoid hammering SQS if no messages are available.
		time.Sleep(5 * time.Second)
	}
}
