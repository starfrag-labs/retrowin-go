package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
	"github.com/starfrag-lab/retrowin-go/internal/object"
)

// UploadService defines the interface for upload operations.
type UploadService interface {
	Upload(ctx context.Context, cmd *UploadCommand) (*UploadResult, error)
	GetDownloadURL(ctx context.Context, id string) (string, error)
	Delete(ctx context.Context, id string) error
}

// UploadCommand for uploading a file.
type UploadCommand struct {
	SystemID string
	UID      int64
	GID      int64
	Mode     int
	Flags    int
	Bucket   string
	Filename string
	MimeType string
	Reader   io.Reader
	Size     int64
}

// UploadResult contains the created inode and object.
type UploadResult struct {
	Inode *inode.Inode
	Object *object.Object
}

// ObjectRef is stored in inode content to reference the Object entity.
type ObjectRef struct {
	ObjectID string `json:"object_id"`
}

type service struct {
	inodeSvc  inode.InodeService
	objectSvc object.ObjectService
}

// NewService creates a new upload service.
func NewService(inodeSvc inode.InodeService, objectSvc object.ObjectService) UploadService {
	return &service{
		inodeSvc:  inodeSvc,
		objectSvc: objectSvc,
	}
}

func (s *service) Upload(ctx context.Context, cmd *UploadCommand) (*UploadResult, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}

	// Create object: streams to storage + creates DB record (atomic)
	// Storage key = inode ID (will be set after inode creation)
	// For now, use system_id/filename as storage key
	storageKey := fmt.Sprintf("%s/%s", cmd.SystemID, cmd.Filename)

	obj, err := s.objectSvc.Create(ctx, &object.CreateCommand{
		Bucket:     cmd.Bucket,
		SystemID:   cmd.SystemID,
		StorageKey: storageKey,
		Reader:     cmd.Reader,
		Size:       cmd.Size,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create object: %w", err)
	}

	// Store Object ID in inode content
	ref := &ObjectRef{ObjectID: obj.ID()}
	refBytes, err := json.Marshal(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal object ref: %w", err)
	}

	// Create inode with object reference
	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeObject | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	}

	createdInode, err := s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
		UID:      cmd.UID,
		GID:      cmd.GID,
		Flags:    cmd.Flags,
		Content:  refBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create inode: %w", err)
	}

	return &UploadResult{
		Inode:  createdInode,
		Object: obj,
	}, nil
}

func (s *service) GetDownloadURL(ctx context.Context, id string) (string, error) {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	var ref ObjectRef
	if err := json.Unmarshal(in.Content(), &ref); err != nil {
		return "", fmt.Errorf("failed to parse object ref: %w", err)
	}
	if ref.ObjectID == "" {
		return "", errors.BadRequest("inode has no object reference")
	}

	return s.objectSvc.GetDownloadURL(ctx, ref.ObjectID)
}

func (s *service) Delete(ctx context.Context, id string) error {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete object (atomic: deletes from storage + DB)
	var ref ObjectRef
	if err := json.Unmarshal(in.Content(), &ref); err == nil && ref.ObjectID != "" {
		if err := s.objectSvc.Delete(ctx, ref.ObjectID); err != nil {
			return fmt.Errorf("failed to delete object: %w", err)
		}
	}

	return s.inodeSvc.Delete(ctx, id)
}
