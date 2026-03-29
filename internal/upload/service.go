package upload

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/starfrag-lab/retrowin-go/internal/fs"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
	"github.com/starfrag-lab/retrowin-go/internal/storage"
)

// Service defines the interface for upload operations.
type Service interface {
	// GetUploadURL generates a presigned upload URL.
	GetUploadURL(ctx context.Context, inodeID int64) (*UploadURL, error)

	// CompleteUpload marks an upload as complete and updates file metadata.
	CompleteUpload(ctx context.Context, inodeID int64, byteSize int64) (*fs.File, error)

	// GetStreamURL generates a presigned download URL.
	GetStreamURL(ctx context.Context, inodeID int64) (*StreamURL, error)
}

// Errors
var (
	ErrCannotUploadDirectory = errors.New("cannot upload content to a directory")
	ErrCannotStreamDirectory = errors.New("cannot stream a directory")
	ErrFileNotFound          = errors.New("file not found")
	ErrContentNotFound       = errors.New("file content not found in storage")
)

type service struct {
	fsSvc   fs.Service
	storage storage.Storage
}

// NewService creates a new upload service.
func NewService(fsSvc fs.Service, storage storage.Storage) Service {
	return &service{
		fsSvc:   fsSvc,
		storage: storage,
	}
}

// GetUploadURL generates a presigned upload URL for a file.
func (s *service) GetUploadURL(ctx context.Context, inodeID int64) (*UploadURL, error) {
	f, err := s.fsSvc.GetByID(ctx, inodeID)
	if err != nil {
		return nil, err
	}

	if f.FileType != inode.FileTypeRegular {
		return nil, ErrCannotUploadDirectory
	}

	storageKey := s.getStorageKey(f.OwnerUID, inodeID)

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
func (s *service) CompleteUpload(ctx context.Context, inodeID int64, byteSize int64) (*fs.File, error) {
	f, err := s.fsSvc.GetByID(ctx, inodeID)
	if err != nil {
		return nil, err
	}

	storageKey := s.getStorageKey(f.OwnerUID, inodeID)
	exists, err := s.storage.ObjectExists(ctx, storageKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check object existence: %w", err)
	}
	if !exists {
		return nil, ErrContentNotFound
	}

	if byteSize == 0 {
		size, err := s.storage.GetObjectSize(ctx, storageKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get object size: %w", err)
		}
		byteSize = size
	}

	if err := s.fsSvc.UpdateByteSize(ctx, inodeID, byteSize); err != nil {
		return nil, fmt.Errorf("failed to update file size: %w", err)
	}

	return s.fsSvc.GetByID(ctx, inodeID)
}

// GetStreamURL generates a presigned download URL for a file.
func (s *service) GetStreamURL(ctx context.Context, inodeID int64) (*StreamURL, error) {
	f, err := s.fsSvc.GetByID(ctx, inodeID)
	if err != nil {
		return nil, err
	}

	if f.FileType != inode.FileTypeRegular {
		return nil, ErrCannotStreamDirectory
	}

	if f.ByteSize == 0 {
		return nil, ErrContentNotFound
	}

	storageKey := s.getStorageKey(f.OwnerUID, inodeID)

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
func (s *service) getStorageKey(ownerUID string, inodeID int64) string {
	return fmt.Sprintf("users/%s/files/%d", ownerUID, inodeID)
}
