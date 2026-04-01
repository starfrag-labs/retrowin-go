// Package memory provides an in-memory storage implementation for testing.
package memory

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	appconfig "github.com/starfrag-lab/retrowin-go/internal/config"
	"github.com/starfrag-lab/retrowin-go/internal/core/object"
	apperrors "github.com/starfrag-lab/retrowin-go/internal/errors"
)

// MemoryStorage implements object.Storage using in-memory storage.
// It generates mock presigned URLs suitable for testing.
type MemoryStorage struct {
	mu            sync.RWMutex
	objects       map[string][]byte  // key -> data
	objectSizes   map[string]int64   // key -> size
	defaultBucket string
	baseURL       string
}

// New creates a new in-memory storage instance.
func New(cfg *appconfig.StorageConfig) (object.Storage, error) {
	if cfg == nil {
		return nil, fmt.Errorf("storage config is required")
	}

	return &MemoryStorage{
		objects:       make(map[string][]byte),
		objectSizes:   make(map[string]int64),
		defaultBucket: cfg.Bucket,
		baseURL:       "http://memory-storage.local",
	}, nil
}

func (s *MemoryStorage) resolveBucket(bucket string) string {
	if bucket != "" {
		return bucket
	}
	return s.defaultBucket
}

func (s *MemoryStorage) objectKey(bucket, key string) string {
	return bucket + "/" + key
}

// PutObject stores data in memory.
func (s *MemoryStorage) PutObject(ctx context.Context, bucket string, key string, reader io.Reader, size int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	fullKey := s.objectKey(s.resolveBucket(bucket), key)
	s.objects[fullKey] = data
	s.objectSizes[fullKey] = size

	return nil
}

// GetPresignedDownloadURL generates a mock presigned URL for download.
// The URL contains the bucket and key information for test verification.
func (s *MemoryStorage) GetPresignedDownloadURL(ctx context.Context, bucket string, key string, expiry time.Duration) (string, error) {
	resolvedBucket := s.resolveBucket(bucket)
	fullKey := s.objectKey(resolvedBucket, key)

	// Generate a mock URL with embedded information
	token := uuid.New().String()
	return fmt.Sprintf("%s/download/%s?bucket=%s&key=%s&token=%s&expires=%d",
		s.baseURL, fullKey, resolvedBucket, key, token, expiry.Milliseconds()), nil
}

// GetPresignedUploadURL generates a mock presigned URL for upload.
// The URL contains the bucket and key information for test verification.
// Auto-stores a placeholder object so CompleteUpload can verify existence.
func (s *MemoryStorage) GetPresignedUploadURL(ctx context.Context, bucket string, key string, contentType string, size int64, expiry time.Duration) (string, error) {
	resolvedBucket := s.resolveBucket(bucket)
	fullKey := s.objectKey(resolvedBucket, key)

	// Auto-store placeholder to simulate successful client upload
	s.mu.Lock()
	s.objects[fullKey] = []byte{}
	s.objectSizes[fullKey] = size
	s.mu.Unlock()

	// Generate a mock URL with embedded information
	token := uuid.New().String()
	return fmt.Sprintf("%s/upload/%s?bucket=%s&key=%s&content_type=%s&size=%d&token=%s&expires=%d",
		s.baseURL, fullKey, resolvedBucket, key, contentType, size, token, expiry.Milliseconds()), nil
}

// DeleteObject removes an object from memory.
func (s *MemoryStorage) DeleteObject(ctx context.Context, bucket string, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fullKey := s.objectKey(s.resolveBucket(bucket), key)
	delete(s.objects, fullKey)
	delete(s.objectSizes, fullKey)

	return nil
}

// ObjectExists checks if an object exists in memory.
func (s *MemoryStorage) ObjectExists(ctx context.Context, bucket string, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fullKey := s.objectKey(s.resolveBucket(bucket), key)
	_, exists := s.objects[fullKey]
	return exists, nil
}

// GetObjectSize returns the size of an object.
func (s *MemoryStorage) GetObjectSize(ctx context.Context, bucket string, key string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fullKey := s.objectKey(s.resolveBucket(bucket), key)
	size, exists := s.objectSizes[fullKey]
	if !exists {
		return 0, apperrors.NotFound("object not found")
	}
	return size, nil
}

// GetObject retrieves raw object data (helper for testing).
func (s *MemoryStorage) GetObject(bucket, key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fullKey := s.objectKey(s.resolveBucket(bucket), key)
	data, exists := s.objects[fullKey]
	return data, exists
}

// Clear removes all objects (helper for testing).
func (s *MemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.objects = make(map[string][]byte)
	s.objectSizes = make(map[string]int64)
}