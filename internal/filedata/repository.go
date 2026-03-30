package filedata

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for file data access.
// This is an internal repository, used only by the filedata service.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*FileData, error)
	GetByInodeID(ctx context.Context, client *ent.Client, inodeID int64) (*FileData, error)
	Update(ctx context.Context, client *ent.Client, params *UpdateParams) error
	Delete(ctx context.Context, client *ent.Client, inodeID int64) error
}

// CreateParams for creating file data (repository layer).
type CreateParams struct {
	InodeID     int64
	StorageType StorageType
	Location    string
	Checksum    *string
}

// UpdateParams for updating file data (repository layer).
type UpdateParams struct {
	InodeID     int64
	StorageType *StorageType
	Location    *string
	Checksum    *string
}
