package object

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for file storage backends.
type Storage interface {
	// PutObject streams data directly to storage. Returns the uploaded size.
	PutObject(ctx context.Context, bucket string, key string, reader io.Reader, size int64) error
	// GetPresignedDownloadURL generates a presigned URL for direct client download.
	GetPresignedDownloadURL(ctx context.Context, bucket string, key string, expiry time.Duration) (string, error)
	DeleteObject(ctx context.Context, bucket string, key string) error
	ObjectExists(ctx context.Context, bucket string, key string) (bool, error)
	GetObjectSize(ctx context.Context, bucket string, key string) (int64, error)
}

const (
	DefaultDownloadExpiry = 1 * time.Hour
)
