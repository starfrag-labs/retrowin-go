package upload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
	"github.com/starfrag-lab/retrowin-go/internal/inode/content"
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
	Inode  *inode.Inode
	Object *object.Object
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

	storageKey := fmt.Sprintf("%s/%s", cmd.SystemID, cmd.Filename)

	// Create object: streams to storage + creates DB record (atomic)
	obj, err := s.objectSvc.Create(ctx, &object.CreateCommand{
		Bucket:     cmd.Bucket,
		SystemID:   cmd.SystemID,
		StorageKey: storageKey,
		Reader:     cmd.Reader,
		Size:       cmd.Size,
	})
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to create object")
	}

	// Store ObjectContent in inode content
	c := &content.ObjectContent{ObjectID: obj.ID()}
	cBytes, err := json.Marshal(c)
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to marshal object content")
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
		Content:  cBytes,
	})
	if err != nil {
		return nil, errors.WrapInternal(err, "failed to create inode")
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

	var c content.ObjectContent
	if err := json.Unmarshal(in.Content(), &c); err != nil {
		return "", errors.WrapInternal(err, "failed to parse object content")
	}
	if c.ObjectID == "" {
		return "", errors.BadRequest("inode has no object reference")
	}

	return s.objectSvc.GetDownloadURL(ctx, c.ObjectID)
}

func (s *service) Delete(ctx context.Context, id string) error {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete object (atomic: deletes from storage + DB)
	var c content.ObjectContent
	if err := json.Unmarshal(in.Content(), &c); err == nil && c.ObjectID != "" {
		if err := s.objectSvc.Delete(ctx, c.ObjectID); err != nil {
			return errors.WrapInternal(err, "failed to delete object")
		}
	}

	return s.inodeSvc.Delete(ctx, id)
}
