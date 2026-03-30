package upload

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
	"github.com/starfrag-lab/retrowin-go/internal/storage"
)

const (
	DefaultUploadExpiry   = 15 * 60 // 15 minutes in seconds
	DefaultDownloadExpiry = 3600    // 1 hour in seconds
)

// Service defines the interface for upload operations.
type Service interface {
	Upload(ctx context.Context, cmd *UploadCommand) (*UploadResult, error)
	GetDownloadURL(ctx context.Context, id int64) (string, error)
	Delete(ctx context.Context, id int64) error
}

// UploadCommand for initiating a file upload.
type UploadCommand struct {
	SystemID string
	UID      int64
	GID      int64
	Mode     int
	Flags    int
	Filename string
	MimeType string
}

// UploadResult contains the upload URL and created inode.
type UploadResult struct {
	Inode     *inode.Inode
	UploadURL string
}

// StorageMetadata is stored in inode content for externally-stored files.
type StorageMetadata struct {
	StorageType string `json:"storage_type"`
	Key         string `json:"key"`
	Bucket      string `json:"bucket,omitempty"`
	Filename    string `json:"filename,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
}

type service struct {
	inodeSvc inode.Service
	storage  storage.Storage
}

// NewService creates a new upload service.
func NewService(inodeSvc inode.Service, storage storage.Storage) Service {
	return &service{
		inodeSvc: inodeSvc,
		storage:  storage,
	}
}

func (s *service) Upload(ctx context.Context, cmd *UploadCommand) (*UploadResult, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}
	if cmd.Filename == "" {
		return nil, errors.BadRequest("filename is required")
	}

	// Generate storage key
	key := fmt.Sprintf("%s/%s", cmd.SystemID, cmd.Filename)

	// Generate presigned upload URL
	uploadURL, err := s.storage.GetPresignedUploadURL(ctx, key, DefaultUploadExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	// Create storage metadata
	meta := &StorageMetadata{
		StorageType: "s3",
		Key:         key,
		Filename:    cmd.Filename,
		MimeType:    cmd.MimeType,
	}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal storage metadata: %w", err)
	}

	// Create inode with storage metadata as content
	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeRegular | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	}

	createdInode, err := s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
		UID:      cmd.UID,
		GID:      cmd.GID,
		Flags:    cmd.Flags,
		Content:  metaBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create inode: %w", err)
	}

	return &UploadResult{
		Inode:     createdInode,
		UploadURL: uploadURL,
	}, nil
}

func (s *service) GetDownloadURL(ctx context.Context, id int64) (string, error) {
	inode, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	var meta StorageMetadata
	if err := json.Unmarshal(inode.Content(), &meta); err != nil {
		return "", fmt.Errorf("failed to parse storage metadata: %w", err)
	}

	if meta.Key == "" {
		return "", errors.BadRequest("inode has no storage key")
	}

	url, err := s.storage.GetPresignedDownloadURL(ctx, meta.Key, DefaultDownloadExpiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return url, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	inode, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Try to delete from external storage
	var meta StorageMetadata
	if err := json.Unmarshal(inode.Content(), &meta); err == nil && meta.Key != "" {
		if err := s.storage.DeleteObject(ctx, meta.Key); err != nil {
			return fmt.Errorf("failed to delete storage object: %w", err)
		}
	}

	return s.inodeSvc.Delete(ctx, id)
}
