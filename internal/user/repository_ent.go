package user

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/user"
)

// EntRepository implements Repository using Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a new EntRepository.
func NewEntRepository(client *ent.Client) Repository {
	return &EntRepository{client: client}
}

// Create creates a new user.
func (r *EntRepository) Create(ctx context.Context, provider, providerID string) (*User, error) {
	entUser, err := r.client.User.
		Create().
		SetProvider(provider).
		SetProviderID(providerID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return fromEntUser(entUser), nil
}

// GetByID retrieves a user by ID.
func (r *EntRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	entUser, err := r.client.User.
		Query().
		Where(user.ID(int(id))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return fromEntUser(entUser), nil
}

// GetByProvider retrieves a user by provider and provider ID.
func (r *EntRepository) GetByProvider(ctx context.Context, provider, providerID string) (*User, error) {
	entUser, err := r.client.User.
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
	return fromEntUser(entUser), nil
}

// Delete deletes a user by ID.
func (r *EntRepository) Delete(ctx context.Context, id int64) error {
	return r.client.User.
		DeleteOneID(int(id)).
		Exec(ctx)
}

// ExistsByProvider checks if a user exists by provider and provider ID.
func (r *EntRepository) ExistsByProvider(ctx context.Context, provider, providerID string) (bool, error) {
	return r.client.User.
		Query().
		Where(
			user.Provider(provider),
			user.ProviderID(providerID),
		).
		Exist(ctx)
}

func fromEntUser(e *ent.User) *User {
	return &User{
		ID:         int64(e.ID),
		Provider:   e.Provider,
		ProviderID: e.ProviderID,
		JoinDate:   e.JoinDate,
		CreatedAt:  e.CreateTime,
		UpdatedAt:  e.UpdateTime,
	}
}
