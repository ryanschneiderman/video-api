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
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	metrics.RegisterMetrics()

	go func() {
		http.Handle("/metrics", metrics.MetricsHandler())
		log.Println("Starting metrics server on :9090")
		log.Fatal(http.ListenAndServe(":9090", nil))
	}()

	ctx := context.Background()
	myApp, err := app.InitializeApp(ctx)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	processor := worker.NewProcessor(myApp)

	// Start polling SQS continuously
	log.Println("Starting SQS worker...")
	for {
		start := time.Now()

		if err := processor.ProcessMessages(ctx); err != nil {
			log.Printf("Error processing messages: %v", err)
			metrics.WorkerErrors.Inc()
		}

		duration := time.Since(start).Seconds()
		metrics.WorkerProcessingDuration.Observe(duration)

		// Sleep briefly to avoid hammering SQS if no messages are available.
		time.Sleep(5 * time.Second)
	}
}
