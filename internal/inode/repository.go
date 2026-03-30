package inode

import (
	"context"
	"time"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// InodeRepository defines the interface for inode data access.
type InodeRepository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Inode, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*Inode, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Inode, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Inode, error)
	UpdateLinkCount(ctx context.Context, client *ent.Client, id int64, delta int) error
}

// CreateParams for creating a new inode (repository layer).
type CreateParams struct {
	SystemID string
	Mode     int
	UID      int64
	GID      int64
	Flags    int
	Content  []byte
}

// UpdateParams for updating an inode (repository layer).
type UpdateParams struct {
	ID    int64
	Mode  *int
	UID   *int64
	GID   *int64
	Size  *int64
	Flags *int
	Atime *time.Time
	Mtime *time.Time
	Ctime *time.Time
}

// QueryFilter for querying inodes (repository layer).
type QueryFilter struct {
	ID       *int64
	SystemID *string
	UID      *int64
	GID      *int64
	Mode     *int
}
