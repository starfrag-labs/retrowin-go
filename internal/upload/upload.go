package upload

import (
	"time"
)

// UploadURL contains presigned upload URL information.
type UploadURL struct {
	UploadURL string    `json:"uploadUrl"`
	Key       string    `json:"key"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// StreamURL contains presigned download URL information.
type StreamURL struct {
	DownloadURL string    `json:"downloadUrl"`
	Key         string    `json:"key"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// Default expiry durations.
const (
	DefaultUploadExpiry = 15 * time.Minute
	DefaultStreamExpiry = 1 * time.Hour
)
