package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDynamoDBClient struct {
	mock.Mock
}

func (m *mockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *mockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func TestNewDB(t *testing.T) {

	t.Skip("Skipping test as it requires AWS credentials")
	

	ctx := context.Background()
	tableName := "test-table"
	
	db, err := NewDB(ctx, tableName)
	
	assert.NoError(t, err)
	assert.NotNil(t, db)
	assert.Equal(t, tableName, db.TableName)
	assert.NotNil(t, db.Client)
}

func TestPutVideo(t *testing.T) {
	mockClient := new(mockDynamoDBClient)
	db := &DB{
		Client:    mockClient,
		TableName: "test-table",
	}
	
	ctx := context.Background()
	video := Video{
		VideoID:     "test-id",
		Title:       "Test Video",
		Description: "Test Description",
		URL:         "https://example.com/video",
		Tags:        []string{"test", "video"},
		UploadDate:  time.Now(),
	}
	

	mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput")).
		Return(&dynamodb.PutItemOutput{}, nil)
	

	err := db.PutVideo(ctx, video)
	
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestPutVideoError(t *testing.T) {
	mockClient := new(mockDynamoDBClient)
	db := &DB{
		Client:    mockClient,
		TableName: "test-table",
	}
	
	ctx := context.Background()
	video := Video{
		VideoID:     "test-id",
		Title:       "Test Video",
		Description: "Test Description",
		URL:         "https://example.com/video",
		Tags:        []string{"test", "video"},
		UploadDate:  time.Now(),
	}
	

	mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput")).
		Return(&dynamodb.PutItemOutput{}, errors.New("DynamoDB error"))
	

	err := db.PutVideo(ctx, video)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to put item in DynamoDB")
	mockClient.AssertExpectations(t)
}

func TestGetVideoById(t *testing.T) {
	mockClient := new(mockDynamoDBClient)
	db := &DB{
		Client:    mockClient,
		TableName: "test-table",
	}
	
	ctx := context.Background()
	videoID := "test-id"
	uploadTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	

	getItemOutput := &dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"video_id":    &types.AttributeValueMemberS{Value: videoID},
			"title":       &types.AttributeValueMemberS{Value: "Test Video"},
			"description": &types.AttributeValueMemberS{Value: "Test Description"},
			"url":         &types.AttributeValueMemberS{Value: "https://example.com/video"},
			"tags":        &types.AttributeValueMemberL{Value: []types.AttributeValue{
				&types.AttributeValueMemberS{Value: "test"},
				&types.AttributeValueMemberS{Value: "video"},
			}},
			"upload_date": &types.AttributeValueMemberS{Value: uploadTime.Format(time.RFC3339)},
		},
	}
	

	mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(getItemOutput, nil)
	

	video, err := db.GetVideoById(ctx, videoID)
	
	assert.NoError(t, err)
	assert.NotNil(t, video)
	assert.Equal(t, videoID, video.VideoID)
	assert.Equal(t, "Test Video", video.Title)
	assert.Equal(t, "Test Description", video.Description)
	assert.Equal(t, "https://example.com/video", video.URL)
	assert.Equal(t, []string{"test", "video"}, video.Tags)
	mockClient.AssertExpectations(t)
}

func TestGetVideoByIdNotFound(t *testing.T) {
	mockClient := new(mockDynamoDBClient)
	db := &DB{
		Client:    mockClient,
		TableName: "test-table",
	}
	
	ctx := context.Background()
	videoID := "nonexistent-id"
	

	mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(&dynamodb.GetItemOutput{}, nil)
	

	video, err := db.GetVideoById(ctx, videoID)
	
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrVideoNotFound))
	assert.Nil(t, video)
	mockClient.AssertExpectations(t)
}

func TestGetVideoByIdError(t *testing.T) {
	mockClient := new(mockDynamoDBClient)
	db := &DB{
		Client:    mockClient,
		TableName: "test-table",
	}
	
	ctx := context.Background()
	videoID := "test-id"
	

	mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput")).
		Return(&dynamodb.GetItemOutput{}, errors.New("DynamoDB error"))
	

	video, err := db.GetVideoById(ctx, videoID)
	
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get item from DynamoDB")

	assert.Nil(t, video)
	mockClient.AssertExpectations(t)
}