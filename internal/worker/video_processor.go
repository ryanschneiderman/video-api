package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/ryanschneiderman/video-api/internal/app"
	"github.com/ryanschneiderman/video-api/internal/db"
)

type Processor struct {
	SQSClient *sqs.Client
	S3Client *s3.Client
	S3Bucket string
	QueueURL  string
	DB       *db.DB
}

type SQSMessage struct {
	VideoID  string `json:"video_id"`
	Filename string `json:"filename"`
	EventType string `json:"event_type,omitempty"`
}

func NewProcessor(app *app.App) *Processor {
	return &Processor{
		SQSClient: app.SQSClient,
		S3Client: app.S3Client,
		S3Bucket: app.S3Bucket,
		QueueURL:  app.QueueURL,
		DB:        app.DB,
	}
}

func (p *Processor) ProcessMessages(ctx context.Context) error {
	output, err := p.SQSClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(p.QueueURL),
		MaxNumberOfMessages: 5, // Processes 5 messages concurrently, increase this value to increase throughput
		WaitTimeSeconds:     10,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeName("ApproximateReceiveCount"), 
		},
	})
	if err != nil {
		return fmt.Errorf("failed to receive messages: %w", err)
	}

	if len(output.Messages) == 0 {
		log.Println("No messages received.")
		return nil
	}

	for _, msg := range output.Messages {
		go func(msg types.Message) {
			if err := p.HandleMessage(ctx, &msg); err != nil {
				log.Printf("Error processing message: %v", err)
			}
		}(msg)
	}

	return nil
}

func (p *Processor) HandleMessage(ctx context.Context, msg *types.Message) error {
	var sqsMsg SQSMessage
	if err := json.Unmarshal([]byte(aws.ToString(msg.Body)), &sqsMsg); err != nil {
		log.Printf("Failed to parse SQS message JSON: %v", err)
		return err
	}

	receiveCount := msg.Attributes["ApproximateReceiveCount"]
	log.Printf("Processing video_id: %s, filename: %s, receive count: %s",
			sqsMsg.VideoID, sqsMsg.Filename, receiveCount)

	if err := p.ProcessVideo(ctx, sqsMsg.VideoID, sqsMsg.Filename); err != nil {
		log.Printf("Error processing video %s: %v", sqsMsg.VideoID, err)
		return err
	}

	_, err := p.SQSClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(p.QueueURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		log.Printf("Failed to delete message for videoID %s: %v", sqsMsg.VideoID, err)
		return err
	}

	log.Printf("Message processed and deleted for videoID: %s", sqsMsg.VideoID)
	return nil
}

func (p *Processor) ProcessVideo(ctx context.Context, videoID string, filename string) error {
	localInputFile := fmt.Sprintf("/tmp/%s", filename)
	localOutputFile := fmt.Sprintf("/tmp/%s-transcoded.mp4", videoID)

	err := downloadFromS3(ctx, p.S3Client, p.S3Bucket, filename, localInputFile)
	if err != nil {
		return fmt.Errorf("failed to download file from S3: %w", err)
	}
	log.Printf("Downloaded %s from S3 to %s", filename, localInputFile)

	err = transcodeVideo(localInputFile, localOutputFile)
	if err != nil {
		return fmt.Errorf("failed to transcode video: %w", err)
	}
	log.Printf("Transcoding complete: %s", localOutputFile)

	aiResult, err := simulateAIInference(localOutputFile)
	if err != nil {
		return fmt.Errorf("AI inference failed: %w", err)
	}
	log.Printf("AI Inference result: %s", aiResult)

	// TODO: this does a PUT video, might want to keep the original UploadDate
	updatedRecord := db.Video{
		VideoID:     videoID,
		Title:       filename,
		Description: fmt.Sprintf("Transcoded and processed: %s", aiResult),
		URL:         fmt.Sprintf("https://%s.s3.amazonaws.com/%s", p.S3Bucket, filename),
		Tags:        []string{"transcoded", "ai-processed"},
		UploadDate:  time.Now(), 
	}
	if err := p.DB.PutVideo(ctx, updatedRecord); err != nil {
		return fmt.Errorf("failed to update video metadata: %w", err)
	}
	log.Printf("Updated video metadata in DynamoDB for videoID: %s", videoID)

	os.Remove(localInputFile)
	os.Remove(localOutputFile)
	return nil
}

func simulateAIInference(videoPath string) (string, error) {
	time.Sleep(2 * time.Second)
	return "This video appears to contain outdoor sports action.", nil
}

func downloadFromS3(ctx context.Context, s3Client *s3.Client, bucket, s3Key, localPath string) error {
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	resp, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save downloaded file: %w", err)
	}

	return nil
}

func transcodeVideo(inputFile, outputFile string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", inputFile, "-vcodec", "libx264", "-acodec", "aac", outputFile)

	output, err := cmd.CombinedOutput()

	if err != nil {
		log.Printf("FFmpeg failed: %v", err)
		log.Printf("FFmpeg Output:\n%s", string(output))
		return fmt.Errorf("failed to transcode video: %w", err)
	}

	return nil
}
