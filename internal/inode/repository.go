package inode

import (
	"context"
	"time"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for inode data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Inode, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*Inode, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Inode, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Inode, error)
	UpdateLinkCount(ctx context.Context, client *ent.Client, id int64, delta int16) error
}

// CreateParams for creating a new inode (repository layer).
type CreateParams struct {
	SystemID   *int64
	FileType   FileType
	OwnerUID   string
	OwnerGID   string
	PermOwner  string
	PermGroup  string
	PermOthers string
	IsSystem   bool
	SystemType *string
}

// UpdateParams for updating an inode (repository layer).
type UpdateParams struct {
	ID         int64
	ByteSize   *int64
	PermOwner  *string
	PermGroup  *string
	PermOthers *string
	LinkCount  *int16
	AccessedAt *time.Time
}

// QueryFilter for querying inodes (repository layer).
type QueryFilter struct {
	ID         *int64
	SystemID   *int64
	OwnerUID   *string
	FileType   *FileType
	IsSystem   *bool
	SystemType *string
}
