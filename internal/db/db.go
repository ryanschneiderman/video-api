package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	ErrVideoNotFound = errors.New("video not found")
	ErrInvalidInput  = errors.New("invalid input")
)

type Video struct {
	VideoID     string    `dynamodbav:"video_id"`
	Title       string    `dynamodbav:"title"`
	Description string    `dynamodbav:"description"`
	URL         string    `dynamodbav:"url"`
	Tags        []string  `dynamodbav:"tags"`
	UploadDate  time.Time `dynamodbav:"upload_date"`
}

type DynamoDBClient interface {
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)

}

type DB struct {
	Client    DynamoDBClient
	TableName string
}

func NewDB(ctx context.Context, tableName string) (*DB, error) {
	if tableName == "" {
		return nil, fmt.Errorf("%w: table name cannot be empty", ErrInvalidInput)
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DB{
		Client:    client,
		TableName: tableName,
	}, nil
}

func (db *DB) PutVideo(ctx context.Context, video Video) error {

	if video.VideoID == "" {
		return fmt.Errorf("%w: video ID cannot be empty", ErrInvalidInput)
	}

	av, err := attributevalue.MarshalMap(video)
	if err != nil {
		return fmt.Errorf("failed to marshal video: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(db.TableName),
		Item:      av,
	}

	_, err = db.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item in DynamoDB: %w", err)
	}

	return nil
}

func (db *DB) GetVideoById(ctx context.Context, videoId string) (*Video, error) {
	if videoId == "" {
		return nil, fmt.Errorf("%w: video ID cannot be empty", ErrInvalidInput)
	}

	key := map[string]types.AttributeValue{
		"video_id": &types.AttributeValueMemberS{Value: videoId},
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(db.TableName),
		Key:       key,
	}

	result, err := db.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	if result.Item == nil || len(result.Item) == 0 {
		return nil, ErrVideoNotFound
	}

	var video Video
	err = attributevalue.UnmarshalMap(result.Item, &video)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &video, nil
}