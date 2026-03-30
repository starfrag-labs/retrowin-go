package user

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/user"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

// Create creates a new user.
func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *CreateParams) (*User, error) {
	entUser, err := client.User.
		Create().
		SetProvider(params.Provider).
		SetProviderID(params.ProviderID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return fromEnt(entUser), nil
}

// GetByID retrieves a user by ID.
func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id int64) (*User, error) {
	entUser, err := client.User.
		Query().
		Where(user.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return fromEnt(entUser), nil
}

// GetByUID retrieves a user by UID.
func (r *EntRepository) GetByUID(ctx context.Context, client *ent.Client, uid string) (*User, error) {
	entUser, err := client.User.
		Query().
		Where(user.UID(uid)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return fromEnt(entUser), nil
}

// GetByProvider retrieves a user by provider and provider ID.
func (r *EntRepository) GetByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (*User, error) {
	entUser, err := client.User.
		Query().
		Where(
			user.Provider(provider),
			user.ProviderID(providerID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return fromEnt(entUser), nil
}

// Delete deletes a user by ID.
func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.User.
		DeleteOneID(id).
		Exec(ctx)
}

// ExistsByProvider checks if a user exists by provider and provider ID.
func (r *EntRepository) ExistsByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (bool, error) {
	return client.User.
		Query().
		Where(
			user.Provider(provider),
			user.ProviderID(providerID),
		).
		Exist(ctx)
}

func fromEnt(e *ent.User) *User {
	return NewUser(
		e.ID,
		e.UID,
		e.Provider,
		e.ProviderID,
		e.JoinDate,
		e.CreateTime,
		e.UpdateTime,
	)
}
