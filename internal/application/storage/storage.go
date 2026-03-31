package storage

import (
	"context"
	"encoding/json"

	"github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/core/object"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// StorageService defines the interface for file storage operations.
type StorageService interface {
	// InitiateUpload creates a pending object and returns presigned upload URL.
	InitiateUpload(ctx context.Context, cmd *InitiateUploadCommand) (*object.UploadSession, error)

	// CompleteUpload verifies upload completion and creates the inode.
	CompleteUpload(ctx context.Context, cmd *CompleteUploadCommand) (*UploadResult, error)

	GetDownloadURL(ctx context.Context, id string) (string, error)
}

// InitiateUploadCommand for starting a presigned upload.
type InitiateUploadCommand struct {
	SystemID    string
	ContentType string
	Size        int64
}

// CompleteUploadCommand for finalizing upload after client confirms.
type CompleteUploadCommand struct {
	ObjectID string
	SystemID string
	Mode     int
	Flags    int
}

// UploadResult contains the created inode and object.
type UploadResult struct {
	Inode  *inode.Inode
	Object *object.Object
}

type service struct {
	fsSvc     fs.FsService
	objectSvc object.ObjectService
}

// NewService creates a new storage service.
func NewService(fsSvc fs.FsService, objectSvc object.ObjectService) StorageService {
	return &service{
		fsSvc:     fsSvc,
		objectSvc: objectSvc,
	}
}

// InitiateUpload creates a pending object and returns presigned upload URL.
func (s *service) InitiateUpload(ctx context.Context, cmd *InitiateUploadCommand) (*object.UploadSession, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}

	session, err := s.objectSvc.InitiateUpload(ctx, &object.InitiateUploadCommand{
		SystemID:    cmd.SystemID,
		ContentType: cmd.ContentType,
		Size:        cmd.Size,
	})
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to initiate upload")
	}

	return session, nil
}

// CompleteUpload verifies upload and creates inode.
func (s *service) CompleteUpload(ctx context.Context, cmd *CompleteUploadCommand) (*UploadResult, error) {
	if cmd.ObjectID == "" {
		return nil, errors.BadRequest("object_id is required")
	}
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}

	// Mark object as active (verifies storage existence internally)
	obj, err := s.objectSvc.CompleteUpload(ctx, cmd.ObjectID)
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to complete upload")
	}

	// Create inode with object reference
	return s.createInodeWithObject(ctx, cmd.SystemID, cmd.Mode, cmd.Flags, obj.ID())
}

// createInodeWithObject creates an inode referencing the given object.
func (s *service) createInodeWithObject(ctx context.Context, systemID string, mode int, flags int, objectID string) (*UploadResult, error) {
	// Store ObjectContent in inode content
	c := &content.ObjectContent{ObjectID: objectID}
	cBytes, err := json.Marshal(c)
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to marshal object content")
	}

	// Set default mode if not provided
	if mode == 0 {
		mode = inode.ModeObject | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	}

	createdInode, err := s.fsSvc.CreateFile(ctx, &fs.CreateFileCommand{
		SystemID: systemID,
		Mode:     mode,
		Flags:    flags,
		Content:  cBytes,
	})
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to create inode")
	}

	// Get the object for the result
	obj, err := s.objectSvc.GetByID(ctx, objectID)
	if err != nil {
		obj = nil // Object not critical for result
	}

	return &UploadResult{
		Inode:  createdInode,
		Object: obj,
	}, nil
}

func (s *service) GetDownloadURL(ctx context.Context, id string) (string, error) {
	in, err := s.fsSvc.Get(ctx, id)
	if err != nil {
		return "", err
	}

	var c content.ObjectContent
	if err := json.Unmarshal(in.Content(), &c); err != nil {
		return "", errors.WrapInternal(err, "failed to parse object content")
	}
	if c.ObjectID == "" {
		return "", errors.BadRequest("inode has no object reference")
	}

	return s.objectSvc.GetDownloadURL(ctx, c.ObjectID)
}
