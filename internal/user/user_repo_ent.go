package user

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	entuser "github.com/starfrag-lab/retrowin-go/ent/user"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *CreateParams) (*User, error) {
	entUser, err := client.User.
		Create().
		SetUsername(params.Username).
		SetProvider(params.Provider).
		SetProviderID(params.ProviderID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return fromEnt(entUser), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id string) (*User, error) {
	entUser, err := client.User.
		Query().
		Where(entuser.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return fromEnt(entUser), nil
}

func (r *EntRepository) GetByUsername(ctx context.Context, client *ent.Client, username string) (*User, error) {
	entUser, err := client.User.
		Query().
		Where(entuser.Username(username)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return fromEnt(entUser), nil
}

func (r *EntRepository) GetByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (*User, error) {
	entUser, err := client.User.
		Query().
		Where(
			entuser.Provider(provider),
			entuser.ProviderID(providerID),
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

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id string) error {
	return client.User.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) ExistsByProvider(ctx context.Context, client *ent.Client, provider, providerID string) (bool, error) {
	return client.User.
		Query().
		Where(
			entuser.Provider(provider),
			entuser.ProviderID(providerID),
		).
		Exist(ctx)
}

func fromEnt(e *ent.User) *User {
	return NewUser(
		e.ID,
		e.Username,
		e.Provider,
		e.ProviderID,
		e.JoinDate,
		e.CreateTime,
		e.UpdateTime,
	)
}
