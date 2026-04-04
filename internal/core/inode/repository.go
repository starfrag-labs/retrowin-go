package inode

import (
	"context"
	"time"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// InodeRepository defines the interface for inode data access.
type InodeRepository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Inode, error)
	GetByID(ctx context.Context, client *ent.Client, id string) (*Inode, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, id string) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Inode, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Inode, error)
	UpdateLinkCount(ctx context.Context, client *ent.Client, id string, delta int) error
}

// CreateParams for creating a new inode (repository layer).
type CreateParams struct {
	ID       string
	SystemID string
	Mode     int
	UID      int
	GID      int
	Flags    int
	Content  []byte
}

// UpdateParams for updating an inode (repository layer).
type UpdateParams struct {
	ID      string
	Mode    *int
	UID     *int
	GID     *int
	Size    *int64
	Flags   *int
	Content *[]byte
	Atime   *time.Time
	Mtime   *time.Time
	Ctime   *time.Time
}

// QueryFilter for querying inodes (repository layer).
type QueryFilter struct {
	ID       *string
	SystemID *string
	UID      *int
	GID      *int
	Mode     *int
}
