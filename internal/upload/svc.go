package upload

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/starfrag-lab/retrowin-go/internal/file"
	"github.com/starfrag-lab/retrowin-go/internal/storage"
)

// Service defines the interface for upload operations.
type Service interface {
	// GetUploadURL generates a presigned upload URL.
	GetUploadURL(ctx context.Context, fileKey string) (*UploadURL, error)

	// CompleteUpload marks an upload as complete and updates file metadata.
	CompleteUpload(ctx context.Context, fileKey string, byteSize int64) (*file.File, error)

	// GetStreamURL generates a presigned download URL.
	GetStreamURL(ctx context.Context, fileKey string) (*StreamURL, error)
}

// Errors
var (
	ErrCannotUploadContainer = errors.New("cannot upload content to a container")
	ErrCannotStreamContainer = errors.New("cannot stream a container")
	ErrFileNotFound          = errors.New("file not found")
	ErrContentNotFound       = errors.New("file content not found in storage")
)

type service struct {
	fileSvc file.Service
	storage storage.Storage
}

// NewService creates a new upload service.
func NewService(fileSvc file.Service, storage storage.Storage) Service {
	return &service{
		fileSvc: fileSvc,
		storage: storage,
	}
}

// GetUploadURL generates a presigned upload URL for a file.
func (s *service) GetUploadURL(ctx context.Context, fileKey string) (*UploadURL, error) {
	// Get file info
	f, err := s.fileSvc.Get(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	// Only files can be uploaded, not containers
	if f.Type != file.FileTypeFile {
		return nil, ErrCannotUploadContainer
	}

	// Generate storage key
	storageKey := s.getStorageKey(f.OwnerID, fileKey)

	// Generate presigned upload URL
	uploadURL, err := s.storage.GetPresignedUploadURL(ctx, storageKey, DefaultUploadExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &UploadURL{
		UploadURL: uploadURL,
		Key:       storageKey,
		ExpiresAt: time.Now().Add(DefaultUploadExpiry),
	}, nil
}

// CompleteUpload marks an upload as complete and updates file metadata.
func (s *service) CompleteUpload(ctx context.Context, fileKey string, byteSize int64) (*file.File, error) {
	// Get file info
	f, err := s.fileSvc.Get(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	// Verify object exists in storage
	storageKey := s.getStorageKey(f.OwnerID, fileKey)
	exists, err := s.storage.ObjectExists(ctx, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check object existence: %w", err)
	}
	if !exists {
		return nil, ErrContentNotFound
	}

	// Get actual size from storage if not provided
	if byteSize == 0 {
		size, err := s.storage.GetObjectSize(ctx, storageKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get object size: %w", err)
		}
		byteSize = size
	}

	// Update file with new size
	updateCmd := &file.UpdateCommand{
		ByteSize: &byteSize,
	}
	updatedFile, err := s.fileSvc.Update(ctx, fileKey, updateCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to update file size: %w", err)
	}

	return updatedFile, nil
}

// GetStreamURL generates a presigned download URL for a file.
func (s *service) GetStreamURL(ctx context.Context, fileKey string) (*StreamURL, error) {
	// Get file info
	f, err := s.fileSvc.Get(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	// Only files can be streamed, not containers
	if f.Type != file.FileTypeFile {
		return nil, ErrCannotStreamContainer
	}

	// Verify file has content
	if f.ByteSize == 0 {
		return nil, ErrContentNotFound
	}

	// Generate storage key
	storageKey := s.getStorageKey(f.OwnerID, fileKey)

	// Generate presigned download URL
	downloadURL, err := s.storage.GetPresignedDownloadURL(ctx, storageKey, DefaultStreamExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &StreamURL{
		DownloadURL: downloadURL,
		Key:         storageKey,
		ExpiresAt:   time.Now().Add(DefaultStreamExpiry),
	}, nil
}

// getStorageKey generates a storage key for a file.
func (s *service) getStorageKey(ownerID int64, fileKey string) string {
	return fmt.Sprintf("users/%d/files/%s", ownerID, fileKey)
}
