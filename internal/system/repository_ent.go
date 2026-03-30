package system

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	entsystem "github.com/starfrag-lab/retrowin-go/ent/system"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *CreateParams) (*System, error) {
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
	return fromEnt(entSystem), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id int64) (*System, error) {
	entSystem, err := client.System.Query().
		Where(entsystem.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system: %w", err)
	}
	return fromEnt(entSystem), nil
}

func (r *EntRepository) GetByName(ctx context.Context, client *ent.Client, name string) (*System, error) {
	entSystem, err := client.System.Query().
		Where(entsystem.Name(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system by name: %w", err)
	}
	return fromEnt(entSystem), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, params *UpdateParams) error {
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

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.System.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*System, error) {
	query := client.System.Query()
	query = applyFilter(query, filter)

	entSystems, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find systems: %w", err)
	}
	return fromEntSlice(entSystems), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*System, error) {
	query := client.System.Query()
	query = applyFilter(query, filter)

	entSystem, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find system: %w", err)
	}
	return fromEnt(entSystem), nil
}

func (r *EntRepository) Exists(ctx context.Context, client *ent.Client, filter *QueryFilter) (bool, error) {
	query := client.System.Query()
	query = applyFilter(query, filter)
	return query.Exist(ctx)
}

func applyFilter(query *ent.SystemQuery, filter *QueryFilter) *ent.SystemQuery {
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

func fromEnt(e *ent.System) *System {
	return NewSystem(
		e.ID,
		e.Name,
		e.Description,
		Status(e.Status),
		e.CreateTime,
		e.UpdateTime,
	)
}

func fromEntSlice(systems []*ent.System) []*System {
	result := make([]*System, len(systems))
	for i, e := range systems {
		result[i] = fromEnt(e)
	}
	return result
}
