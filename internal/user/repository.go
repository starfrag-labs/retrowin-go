package user

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for user data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*User, error)
	GetByID(ctx context.Context, client *ent.Client, id int64) (*User, error)
	GetByUID(ctx context.Context, client *ent.Client, uid string) (*User, error)
	GetByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (*User, error)
	Delete(ctx context.Context, client *ent.Client, id int64) error
	ExistsByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (bool, error)
}

// CreateParams for creating a user (repository layer).
type CreateParams struct {
	Provider   string
	ProviderID string
}

// QueryFilter for querying users (repository layer).
type QueryFilter struct {
	ID         *int64
	UID        *string
	Provider   *string
	ProviderID *string
}
