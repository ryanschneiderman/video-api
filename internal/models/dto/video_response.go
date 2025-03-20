package models

type VideoResponse struct {
	VideoID     string                 `json:"videoId"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	URL         string                 `json:"url"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UploadDate  string            `json:"uploadDate,omitempty"`
}