package inode

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for inode data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Inode, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*Inode, error)
	Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error
	Delete(ctx context.Context, client *ent.Client, id int64) error
	Find(ctx context.Context, client *ent.Client, filter Filter) ([]*Inode, error)
	FindOne(ctx context.Context, client *ent.Client, filter Filter) (*Inode, error)
	UpdateLinkCount(ctx context.Context, client *ent.Client, id int64, delta int16) error
}

// FileDataRepository defines the interface for file data access.
// This is an internal repository, not exposed as a service.
type FileDataRepository interface {
	Create(ctx context.Context, client *ent.Client, cmd *CreateFileDataCommand) (*FileData, error)
	GetByInodeID(ctx context.Context, client *ent.Client, inodeID int64) (*FileData, error)
	Update(ctx context.Context, client *ent.Client, cmd *UpdateFileDataCommand) error
	Delete(ctx context.Context, client *ent.Client, inodeID int64) error
}

// CreateFileDataCommand for creating file data.
type CreateFileDataCommand struct {
	InodeID     int64
	StorageType StorageType
	Location    string
	Checksum    *string
}

// UpdateFileDataCommand for updating file data.
type UpdateFileDataCommand struct {
	InodeID     int64
	StorageType *StorageType
	Location    *string
	Checksum    *string
}
