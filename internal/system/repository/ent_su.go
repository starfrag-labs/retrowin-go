package repository

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	entusersystem "github.com/starfrag-lab/retrowin-go/ent/usersystem"
	domain "github.com/starfrag-lab/retrowin-go/internal/system"
)

// EntSystemUserRepository implements domain.SystemUserRepository using Ent.
type EntSystemUserRepository struct{}

// NewSystemUserRepository creates a new EntSystemUserRepository.
func NewSystemUserRepository() domain.SystemUserRepository {
	return &EntSystemUserRepository{}
}

func (r *EntSystemUserRepository) Create(ctx context.Context, client *ent.Client, params *domain.SystemUserCreateParams) (*domain.SystemUser, error) {
	entUserSystem, err := client.UserSystem.Create().
		SetUserID(params.UserID).
		SetSystemID(params.SystemID).
		SetUsername(params.Username).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create system user: %w", err)
	}
	return systemUserFromEnt(entUserSystem), nil
}

func (r *EntSystemUserRepository) GetByID(ctx context.Context, client *ent.Client, id int) (*domain.SystemUser, error) {
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

func (r *EntSystemUserRepository) Find(ctx context.Context, client *ent.Client, filter *domain.SystemUserQueryFilter) ([]*domain.SystemUser, error) {
	query := client.UserSystem.Query()
	query = applySystemUserFilter(query, filter)

	entUserSystems, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find system users: %w", err)
	}
	return systemUserFromEntSlice(entUserSystems), nil
}

func (r *EntSystemUserRepository) FindOne(ctx context.Context, client *ent.Client, filter *domain.SystemUserQueryFilter) (*domain.SystemUser, error) {
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

func applySystemUserFilter(query *ent.UserSystemQuery, filter *domain.SystemUserQueryFilter) *ent.UserSystemQuery {
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

func systemUserFromEnt(e *ent.UserSystem) *domain.SystemUser {
	return domain.NewSystemUser(
		e.ID,
		e.UserID,
		e.SystemID,
		e.Username,
	)
}

func systemUserFromEntSlice(items []*ent.UserSystem) []*domain.SystemUser {
	result := make([]*domain.SystemUser, len(items))
	for i, e := range items {
		result[i] = systemUserFromEnt(e)
	}
	return result
}
