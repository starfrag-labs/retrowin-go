package object

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// ObjectRepository defines the interface for object data access.
type ObjectRepository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Object, error)
	GetByID(ctx context.Context, client *ent.Client, id string) (*Object, error)
	GetByStorageKey(ctx context.Context, client *ent.Client, systemID string, provider string, bucket string, storageKey string) (*Object, error)
	Delete(ctx context.Context, client *ent.Client, id string) error
	Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Object, error)
	FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Object, error)
}

// CreateParams for creating a new object (repository layer).
type CreateParams struct {
	Provider   Provider
	Bucket     string
	SystemID   string
	StorageKey string
}

// QueryFilter for querying objects (repository layer).
type QueryFilter struct {
	ID         *string
	SystemID   *string
	Provider   *string
	Bucket     *string
	StorageKey *string
}
