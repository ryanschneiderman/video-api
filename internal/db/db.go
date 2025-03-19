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

var ErrVideoNotFound = errors.New("video not found")

type Video struct {
	VideoID     string    `dynamodbav:"video_id"`
	Title       string    `dynamodbav:"title"`
	Description string    `dynamodbav:"description"`
	URL         string    `dynamodbav:"url"`
	Tags        []string  `dynamodbav:"tags"`
	UploadDate  time.Time `dynamodbav:"upload_date"`
}

type DB struct {
	Client    *dynamodb.Client
	TableName string
}

func NewDB(ctx context.Context, tableName string) (*DB, error) {
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
	key := map[string]types.AttributeValue{
		"video_id": &types.AttributeValueMemberS{Value: videoId},
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(db.TableName),
		Key:       key,
	}

	result, err := db.Client.GetItem(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("failed to get item from dynamoDb: %w", err)
	}

	if result.Item == nil {
		return nil, ErrVideoNotFound
	}

	var video Video
	err = attributevalue.UnmarshalMap(result.Item, &video)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %w", err)
	}

	return &video, nil
}
