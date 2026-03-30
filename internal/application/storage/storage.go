package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/core/object"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// StorageService defines the interface for file storage operations.
type StorageService interface {
	Upload(ctx context.Context, cmd *UploadCommand) (*UploadResult, error)
	GetDownloadURL(ctx context.Context, id string) (string, error)
}

// UploadCommand for uploading a file.
type UploadCommand struct {
	SystemID string
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

	// Create inode via fs (UID resolved from context internally)
	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeObject | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	}

	createdInode, err := s.fsSvc.CreateFile(ctx, &fs.CreateFileCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
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
