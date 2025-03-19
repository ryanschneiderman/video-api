package app

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/ryanschneiderman/video-api/internal/db"
)

// App holds shared AWS clients and configuration.
type App struct {
	DB       *db.DB        // Our DynamoDB wrapper, not just the raw client.
	S3Client *s3.Client       // S3 client
	SQSClient *sqs.Client     // SQS client
	TableName string          // DynamoDB table name
	S3Bucket  string          // S3 bucket name
	QueueURL  string          // SQS Queue URL
}

// InitializeApp creates and returns an App with all AWS clients initialized.
func InitializeApp(ctx context.Context) (*App, error) {
	// Load AWS config (from env vars, shared config file, etc.)
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	// Create clients
	s3Client := s3.NewFromConfig(cfg)
	sqsClient := sqs.NewFromConfig(cfg)

	// Get environment variables
	tableName := os.Getenv("DYNAMODB_TABLE")
	if tableName == "" {
		return nil, fmt.Errorf("DYNAMODB_TABLE env variable not set")
	}
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET env variable not set")
	}
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		return nil, fmt.Errorf("SQS_QUEUE_URL env variable not set")
	}

	// Initialize our DB wrapper.
    dbWrapper, err := db.NewDB(ctx, tableName)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize DB wrapper: %w", err)
    }

	return &App{
		DB:        dbWrapper,
		S3Client:  s3Client,
		SQSClient: sqsClient,
		TableName: tableName,
		S3Bucket:  bucket,
		QueueURL:  queueURL,
	}, nil
}
