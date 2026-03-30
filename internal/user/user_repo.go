package user

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/ent"
)

// Repository defines the interface for user data access.
type Repository interface {
	Create(ctx context.Context, client *ent.Client, params *CreateParams) (*User, error)
	GetByID(ctx context.Context, client *ent.Client, id string) (*User, error)
	GetByUsername(ctx context.Context, client *ent.Client, username string) (*User, error)
	GetByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (*User, error)
	Delete(ctx context.Context, client *ent.Client, id string) error
	ExistsByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (bool, error)
}

// CreateParams for creating a user (repository layer).
type CreateParams struct {
	Username   string
	Provider   string
	ProviderID string
}
