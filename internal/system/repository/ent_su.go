package repository

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	entusersystem "github.com/starfrag-lab/retrowin-go/ent/usersystem"
	"github.com/starfrag-lab/retrowin-go/internal/core/user"
)

// EntSystemUserRepository implements user.SystemUserRepository using Ent.
type EntSystemUserRepository struct{}

// NewSystemUserRepository creates a new EntSystemUserRepository.
func NewSystemUserRepository() user.SystemUserRepository {
	return &EntSystemUserRepository{}
}

func (r *EntSystemUserRepository) Create(ctx context.Context, client *ent.Client, params *user.CreateParams) (*user.SystemUser, error) {
	entUserSystem, err := client.UserSystem.Create().
		SetUserID(params.UserID).
		SetSystemID(params.SystemID).
		SetUsername(params.Username).
		SetUID(params.UID).
		SetGid(params.GID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create system user: %w", err)
	}
	return systemUserFromEnt(entUserSystem), nil
}

func (r *EntSystemUserRepository) GetByID(ctx context.Context, client *ent.Client, id int) (*user.SystemUser, error) {
	entUserSystem, err := client.UserSystem.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system user: %w", err)
	}
	return systemUserFromEnt(entUserSystem), nil
}

func (r *EntSystemUserRepository) Delete(ctx context.Context, client *ent.Client, id int) error {
	return client.UserSystem.DeleteOneID(id).Exec(ctx)
}

func (r *EntSystemUserRepository) Find(ctx context.Context, client *ent.Client, filter *user.QueryFilter) ([]*user.SystemUser, error) {
	query := client.UserSystem.Query()
	query = applySystemUserFilter(query, filter)

	entUserSystems, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find system users: %w", err)
	}
	return systemUserFromEntSlice(entUserSystems), nil
}

func (r *EntSystemUserRepository) FindOne(ctx context.Context, client *ent.Client, filter *user.QueryFilter) (*user.SystemUser, error) {
	query := client.UserSystem.Query()
	query = applySystemUserFilter(query, filter)

	entUserSystem, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find system user: %w", err)
	}
	return systemUserFromEnt(entUserSystem), nil
}

// GetNextUID returns the next available UID for the system.
// It finds the max UID in the system and returns max+1, or MinUID if no users exist.
func (r *EntSystemUserRepository) GetNextUID(ctx context.Context, client *ent.Client, systemID string) (int, error) {
	// Find max UID in the system
	maxUID, err := client.UserSystem.Query().
		Where(entusersystem.SystemIDEQ(systemID)).
		Aggregate(
			ent.Max(entusersystem.FieldUID),
		).
		Int(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return user.MinUID, nil // No users, start from MinUID
		}
		return 0, fmt.Errorf("failed to get max uid: %w", err)
	}

	nextUID := maxUID + 1
	if nextUID > user.MaxUID {
		return 0, fmt.Errorf("no available uid (max %d reached)", user.MaxUID)
	}

	// Ensure we start from MinUID
	if nextUID < user.MinUID {
		nextUID = user.MinUID
	}

	return nextUID, nil
}

func applySystemUserFilter(query *ent.UserSystemQuery, filter *user.QueryFilter) *ent.UserSystemQuery {
	if filter == nil {
		return query
	}
	if filter.UserID != nil {
		query = query.Where(entusersystem.UserIDEQ(*filter.UserID))
	}
	if filter.SystemID != nil {
		query = query.Where(entusersystem.SystemIDEQ(*filter.SystemID))
	}
	if filter.Username != nil {
		query = query.Where(entusersystem.UsernameEQ(*filter.Username))
	}
	return query
}

func systemUserFromEnt(e *ent.UserSystem) *user.SystemUser {
	return user.NewSystemUser(
		e.ID,
		e.UserID,
		e.SystemID,
		e.Username,
		e.UID,
		e.Gid,
	)
}

func systemUserFromEntSlice(items []*ent.UserSystem) []*user.SystemUser {
	result := make([]*user.SystemUser, len(items))
	for i, e := range items {
		result[i] = systemUserFromEnt(e)
	}
	return result
}
