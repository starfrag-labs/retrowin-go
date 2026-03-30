package fs

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
)

// FsService defines the interface for filesystem operations.
type FsService interface {
	CreateFile(ctx context.Context, cmd *CreateFileCommand) (*inode.Inode, error)
	CreateDirectory(ctx context.Context, cmd *CreateDirectoryCommand) (*inode.Inode, error)
	Get(ctx context.Context, id int64) (*inode.Inode, error)
	UpdateContent(ctx context.Context, cmd *UpdateContentCommand) (*inode.Inode, error)
	UpdateMode(ctx context.Context, cmd *UpdateModeCommand) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter *ListFilter) ([]*inode.Inode, error)
	Copy(ctx context.Context, id int64, systemID string) (*inode.Inode, error)
}

// CreateFileCommand for creating a regular file.
type CreateFileCommand struct {
	SystemID string
	UID      int64
	GID      int64
	Mode     int
	Flags    int
	Content  []byte
}

// CreateDirectoryCommand for creating a directory.
type CreateDirectoryCommand struct {
	SystemID string
	UID      int64
	GID      int64
	Mode     int
	Flags    int
}

// UpdateContentCommand for updating file content.
type UpdateContentCommand struct {
	ID      int64
	Content []byte
}

// UpdateModeCommand for updating file mode (permissions).
type UpdateModeCommand struct {
	ID   int64
	Mode int
}

// ListFilter for listing inodes.
type ListFilter struct {
	SystemID *string
	UID      *int64
}

type service struct {
	inodeSvc inode.InodeService
}

// NewService creates a new filesystem service.
func NewService(inodeSvc inode.InodeService) FsService {
	return &service{inodeSvc: inodeSvc}
}

func (s *service) CreateFile(ctx context.Context, cmd *CreateFileCommand) (*inode.Inode, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}

	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeRegular | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	}

	return s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
		UID:      cmd.UID,
		GID:      cmd.GID,
		Flags:    cmd.Flags,
		Content:  cmd.Content,
	})
}

func (s *service) CreateDirectory(ctx context.Context, cmd *CreateDirectoryCommand) (*inode.Inode, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}

	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherR
	}

	return s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
		UID:      cmd.UID,
		GID:      cmd.GID,
		Flags:    cmd.Flags,
	})
}

func (s *service) Get(ctx context.Context, id int64) (*inode.Inode, error) {
	return s.inodeSvc.GetByID(ctx, id)
}

func (s *service) UpdateContent(ctx context.Context, cmd *UpdateContentCommand) (*inode.Inode, error) {
	// Get current inode to update size
	current, err := s.inodeSvc.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	size := int64(len(cmd.Content))
	if err := s.inodeSvc.Update(ctx, &inode.UpdateCommand{
		ID:   cmd.ID,
		Size: &size,
	}); err != nil {
		return nil, err
	}

	return current, nil
}

func (s *service) UpdateMode(ctx context.Context, cmd *UpdateModeCommand) error {
	return s.inodeSvc.Update(ctx, &inode.UpdateCommand{
		ID:   cmd.ID,
		Mode: &cmd.Mode,
	})
}

func (s *service) Delete(ctx context.Context, id int64) error {
	return s.inodeSvc.Delete(ctx, id)
}

func (s *service) List(ctx context.Context, filter *ListFilter) ([]*inode.Inode, error) {
	f := inode.Filter{}
	if filter.SystemID != nil {
		f = inode.BySystemID(*filter.SystemID)
	}
	if filter.UID != nil {
		f.UID = filter.UID
	}
	return s.inodeSvc.Find(ctx, f)
}

func (s *service) Copy(ctx context.Context, id int64, systemID string) (*inode.Inode, error) {
	original, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: systemID,
		Mode:     original.Mode(),
		UID:      original.UID(),
		GID:      original.GID(),
		Flags:    original.Flags(),
		Content:  original.Content(),
	})
}
