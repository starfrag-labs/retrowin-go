package repository

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	entsystem "github.com/starfrag-lab/retrowin-go/ent/system"
	domain "github.com/starfrag-lab/retrowin-go/internal/system"
)

// EntRepository implements domain.SystemRepository using Ent.
type EntRepository struct{}

// NewRepository creates a new EntRepository.
func NewRepository() domain.SystemRepository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *domain.CreateParams) (*domain.System, error) {
	builder := client.System.Create().
		SetName(params.Name).
		SetStatus(entsystem.Status(params.Status))

	if params.Description != nil {
		builder.SetDescription(*params.Description)
	}

	entSystem, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create system: %w", err)
	}
	return systemFromEnt(entSystem), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id string) (*domain.System, error) {
	entSystem, err := client.System.Query().
		Where(entsystem.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system: %w", err)
	}
	return systemFromEnt(entSystem), nil
}

func (r *EntRepository) GetByName(ctx context.Context, client *ent.Client, name string) (*domain.System, error) {
	entSystem, err := client.System.Query().
		Where(entsystem.Name(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system by name: %w", err)
	}
	return systemFromEnt(entSystem), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, params *domain.UpdateParams) error {
	builder := client.System.UpdateOneID(params.ID)

	if params.Name != nil {
		builder.SetName(*params.Name)
	}
	if params.Description != nil {
		builder.SetDescription(*params.Description)
	}
	if params.Status != nil {
		builder.SetStatus(entsystem.Status(*params.Status))
	}

	return builder.Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id string) error {
	return client.System.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter *domain.QueryFilter) ([]*domain.System, error) {
	query := client.System.Query()
	query = applySystemFilter(query, filter)

	entSystems, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find systems: %w", err)
	}
	return systemFromEntSlice(entSystems), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter *domain.QueryFilter) (*domain.System, error) {
	query := client.System.Query()
	query = applySystemFilter(query, filter)

	entSystem, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find system: %w", err)
	}
	return systemFromEnt(entSystem), nil
}

func (r *EntRepository) Exists(ctx context.Context, client *ent.Client, filter *domain.QueryFilter) (bool, error) {
	query := client.System.Query()
	query = applySystemFilter(query, filter)
	return query.Exist(ctx)
}

func applySystemFilter(query *ent.SystemQuery, filter *domain.QueryFilter) *ent.SystemQuery {
	if filter == nil {
		return query
	}
	if filter.ID != nil {
		query = query.Where(entsystem.ID(*filter.ID))
	}
	if filter.Name != nil {
		query = query.Where(entsystem.Name(*filter.Name))
	}
	if filter.Status != nil {
		query = query.Where(entsystem.StatusEQ(entsystem.Status(*filter.Status)))
	}
	return query
}

func systemFromEnt(e *ent.System) *domain.System {
	return domain.NewSystem(
		e.ID,
		e.Name,
		e.Description,
		domain.Status(e.Status),
		e.CreateTime,
		e.UpdateTime,
	)
}

func systemFromEntSlice(systems []*ent.System) []*domain.System {
	result := make([]*domain.System, len(systems))
	for i, e := range systems {
		result[i] = systemFromEnt(e)
	}
	return result
}
