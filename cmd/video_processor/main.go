package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/ryanschneiderman/video-api/internal/app"
	"github.com/ryanschneiderman/video-api/internal/worker"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize the shared app (AWS clients, configuration, etc.)
	ctx := context.Background()
	myApp, err := app.InitializeApp(ctx)
	if err != nil {
		log.Fatal("Failed to initialize application:", err)
	}

	// Create a new worker processor with the shared dependencies
	processor := worker.NewProcessor(myApp)

	// Start polling SQS continuously
	log.Println("Starting SQS worker...")
	for {
		if err := processor.ProcessMessages(ctx); err != nil {
			log.Printf("Error processing messages: %v", err)
		}
		// Sleep briefly to avoid hammering SQS if no messages
		time.Sleep(5 * time.Second)
	}
}
