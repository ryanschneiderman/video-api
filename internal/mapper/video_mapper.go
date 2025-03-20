package mapper

import (
	"time"

	db "github.com/ryanschneiderman/video-api/internal/db"
	models "github.com/ryanschneiderman/video-api/internal/models/dto"
)

func ToVideoResponse(video *db.Video) *models.VideoResponse {
	return &models.VideoResponse{
		VideoID:     video.VideoID,
		Title:       video.Title,
		Description: video.Description,
		Tags:        video.Tags,
		URL:         video.URL,
		Metadata:    video.Metadata,
		UploadDate:  video.UploadDate.Format(time.RFC3339),
	}
}
