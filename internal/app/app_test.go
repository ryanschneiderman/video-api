package app_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ryanschneiderman/video-api/internal/app"
)

func TestInitializeApp_Success(t *testing.T) {
	// Set necessary environment variables.
	os.Setenv("DYNAMODB_TABLE", "test-table")
	os.Setenv("S3_BUCKET", "test-bucket")
	os.Setenv("SQS_QUEUE_URL", "http://test-queue")
	os.Setenv("AWS_REGION", "us-east-1")
	defer func() {
		os.Unsetenv("DYNAMODB_TABLE")
		os.Unsetenv("S3_BUCKET")
		os.Unsetenv("SQS_QUEUE_URL")
		os.Unsetenv("AWS_REGION")
	}()

	ctx := context.Background()
	a, err := app.InitializeApp(ctx)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if a.TableName != "test-table" {
		t.Errorf("expected TableName to be 'test-table', got: %s", a.TableName)
	}
	if a.S3Bucket != "test-bucket" {
		t.Errorf("expected S3Bucket to be 'test-bucket', got: %s", a.S3Bucket)
	}
	if a.QueueURL != "http://test-queue" {
		t.Errorf("expected QueueURL to be 'http://test-queue', got: %s", a.QueueURL)
	}
}

func TestInitializeApp_MissingEnv(t *testing.T) {
	os.Unsetenv("DYNAMODB_TABLE")
	os.Unsetenv("S3_BUCKET")
	os.Unsetenv("SQS_QUEUE_URL")
	os.Unsetenv("AWS_REGION")

	ctx := context.Background()
	_, err := app.InitializeApp(ctx)
	if err == nil {
		t.Fatal("expected an error due to missing environment variables, got nil")
	}

	if !strings.Contains(err.Error(), "DYNAMODB_TABLE env variable not set") {
		t.Errorf("expected error to mention missing DYNAMODB_TABLE, got: %v", err)
	}
}
