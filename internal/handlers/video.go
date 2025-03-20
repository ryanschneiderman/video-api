package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ryanschneiderman/video-api/internal/app"
	"github.com/ryanschneiderman/video-api/internal/db"
	"github.com/ryanschneiderman/video-api/internal/mapper"
)

type VideoHandler struct {
	DB       *db.DB
	S3Client  *s3.Client
	SQSClient *sqs.Client
	TableName string
	S3Bucket  string
	QueueURL  string
}

func NewVideoHandler(app *app.App) *VideoHandler {
	return &VideoHandler{
		DB:        app.DB,
		S3Client:  app.S3Client,
		SQSClient: app.SQSClient,
		TableName: app.TableName,
		S3Bucket:  app.S3Bucket,
		QueueURL:  app.QueueURL,
	}
}

func (vh *VideoHandler) UploadVideo(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		log.Println("Error retrieving file from request:", err)
		c.JSON(400, gin.H{"error": "Invalid file upload"})
		return
	}
	defer file.Close()

	videoID := uuid.New().String()
	filename := fmt.Sprintf("%s-%s", videoID, header.Filename)
	log.Println("Uploading file:", filename)

	url, s3err := vh.uploadToS3(file, filename)
	if s3err != nil {
		log.Println("Error uploading to S3:", s3err)
		c.JSON(500, gin.H{"error": "Failed to upload video"})
		return
	}

	videoRecord := db.Video{
		VideoID:     videoID,
		Title:       header.Filename,
		Description: "A newly uploaded video",
		URL:         url,
		Metadata: nil,
		Tags:        []string{},
		UploadDate:  time.Now(),
	}

	ctx := context.TODO()
	if err = vh.DB.PutVideo(ctx, videoRecord); err != nil {
		log.Println("Error saving video record:", err)
		c.JSON(500, gin.H{"error": "Failed to save video metadata"})
		return
	}

	sqsErr := vh.sendSQSMessage(videoID, filename)
	if sqsErr != nil {
		log.Println("Error sending SQS message:", sqsErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enqueue processing job"})
		return
	}

	log.Println("Successfully enqueued video processing message")

	c.JSON(201, gin.H{
		"videoId": videoID,
	})
}

func (vh *VideoHandler) GetVideo(c *gin.Context){
	videoId := c.Param("id")
	
	if _, err := uuid.Parse(videoId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid video ID format"})
		return
	}

	ctx := context.TODO() 
	video, err := vh.DB.GetVideoById(ctx, videoId)
	videoResponse := mapper.ToVideoResponse(video)
	if err != nil{
		if errors.Is(err, db.ErrVideoNotFound) {
			log.Printf("Video not found with ID: %s", videoId)
			c.JSON(http.StatusNotFound, gin.H{"error": "Video not found"})
			return
		}
		log.Printf("Failed to get video with ID: %s, error: %v", videoId, err)
		c.JSON(500, gin.H{"error": "Failed to save video metadata"})
	}
	c.JSON(200, videoResponse)
}

func (vh *VideoHandler) uploadToS3(file multipart.File, filename string) (string, error) {
	_, err := vh.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(vh.S3Bucket),
		Key:    aws.String(filename),
		Body:   file,
		ACL:    "private",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", vh.S3Bucket, filename)
	return url, nil
}

func (vh *VideoHandler) sendSQSMessage(videoId string, filename string) error {

	messageBody := fmt.Sprintf(`{"video_id": "%s", "filename": "%s"}`, videoId, filename)

	sqsInput := &sqs.SendMessageInput{
		QueueUrl:    aws.String(vh.QueueURL),
		MessageBody: aws.String(messageBody),
	}

	_, err := vh.SQSClient.SendMessage(context.TODO(), sqsInput)
	if err != nil {
		return fmt.Errorf("failed to send SQS message: %w", err)
	}
	return nil
}
