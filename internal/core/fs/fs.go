package fs

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/core/object"
	"github.com/starfrag-lab/retrowin-go/internal/core/user"
)

// FsService defines the interface for filesystem operations.
type FsService interface {
	CreateFile(ctx context.Context, cmd *CreateFileCommand) (*inode.Inode, error)
	CreateDirectory(ctx context.Context, cmd *CreateDirectoryCommand) (*inode.Inode, error)
	CreateSymlink(ctx context.Context, cmd *CreateSymlinkCommand) (*inode.Inode, error)
	Get(ctx context.Context, id string) (*inode.Inode, error)
	ReadDir(ctx context.Context, id string) ([]content.DirEntry, error)
	Link(ctx context.Context, dirID string, entry content.DirEntry) error
	Unlink(ctx context.Context, dirID string, name string) error
	UpdateContent(ctx context.Context, cmd *UpdateContentCommand) (*inode.Inode, error)
	UpdateMode(ctx context.Context, cmd *UpdateModeCommand) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter *ListFilter) ([]*inode.Inode, error)
	Copy(ctx context.Context, id string, systemID string) (*inode.Inode, error)
}

// CreateFileCommand for creating a regular file.
type CreateFileCommand struct {
	SystemID string
	GID      int
	Mode     int
	Flags    int
	Content  []byte
}

// CreateDirectoryCommand for creating a directory.
type CreateDirectoryCommand struct {
	SystemID string
	GID      int
	Mode     int
	Flags    int
}

// CreateSymlinkCommand for creating a symbolic link.
type CreateSymlinkCommand struct {
	SystemID string
	GID      int
	Mode     int
	Flags    int
	Target   string
}

// UpdateContentCommand for updating file content.
type UpdateContentCommand struct {
	ID      string
	Content []byte
}

// UpdateModeCommand for updating file mode (permissions).
type UpdateModeCommand struct {
	ID   string
	Mode int
}

// ListFilter for listing inodes.
type ListFilter struct {
	SystemID *string
	UID      *int
}

type service struct {
	inodeSvc  inode.InodeService
	objectSvc object.ObjectService
	userSvc   user.UserService
}

// NewService creates a new filesystem service.
func NewService(inodeSvc inode.InodeService, objectSvc object.ObjectService, userSvc user.UserService) FsService {
	return &service{
		inodeSvc:  inodeSvc,
		objectSvc: objectSvc,
		userSvc:   userSvc,
	}
}
