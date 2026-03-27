package storage

import (
	"context"
	"time"
)

// Storage defines the interface for file storage backends.
// Implementations can include S3, MinIO, GCS, etc.
type Storage interface {
	// GetPresignedUploadURL generates a presigned URL for direct client upload.
	GetPresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// GetPresignedDownloadURL generates a presigned URL for direct client download.
	GetPresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// DeleteObject removes an object from storage.
	DeleteObject(ctx context.Context, key string) error

	// ObjectExists checks if an object exists in storage.
	ObjectExists(ctx context.Context, key string) (bool, error)

	// GetObjectSize returns the size of an object in bytes.
	GetObjectSize(ctx context.Context, key string) (int64, error)
}

// StorageType defines the type of storage backend.
type StorageType string

const (
	StorageTypeS3    StorageType = "s3"
	StorageTypeMinIO StorageType = "minio"
	StorageTypeGCS   StorageType = "gcs"
)

// Default presigned URL expiry times
const (
	DefaultUploadExpiry   = 15 * time.Minute
	DefaultDownloadExpiry = 1 * time.Hour
)
