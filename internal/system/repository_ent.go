package system

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/system"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*System, error) {
	builder := client.System.Create().
		SetName(cmd.Name).
		SetStatus(system.Status(cmd.Status))

	if cmd.Description != nil {
		builder.SetDescription(*cmd.Description)
	}

	entSystem, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create system: %w", err)
	}
	return fromEntSystem(entSystem), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id int64) (*System, error) {
	entSystem, err := client.System.Query().
		Where(system.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system: %w", err)
	}
	return fromEntSystem(entSystem), nil
}

func (r *EntRepository) GetByName(ctx context.Context, client *ent.Client, name string) (*System, error) {
	entSystem, err := client.System.Query().
		Where(system.Name(name)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get system by name: %w", err)
	}
	return fromEntSystem(entSystem), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error {
	builder := client.System.UpdateOneID(cmd.ID)

	if cmd.Name != nil {
		builder.SetName(*cmd.Name)
	}
	if cmd.Description != nil {
		builder.SetDescription(*cmd.Description)
	}
	if cmd.Status != nil {
		builder.SetStatus(system.Status(*cmd.Status))
	}

	return builder.Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.System.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter Filter) ([]*System, error) {
	query := client.System.Query()
	query = applyFilter(query, filter)

	entSystems, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find systems: %w", err)
	}
	return fromEntSystems(entSystems), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter Filter) (*System, error) {
	query := client.System.Query()
	query = applyFilter(query, filter)

	entSystem, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find system: %w", err)
	}
	return fromEntSystem(entSystem), nil
}

func (r *EntRepository) Exists(ctx context.Context, client *ent.Client, filter Filter) (bool, error) {
	query := client.System.Query()
	query = applyFilter(query, filter)
	return query.Exist(ctx)
}

func applyFilter(query *ent.SystemQuery, filter Filter) *ent.SystemQuery {
	if filter.ID != nil {
		query = query.Where(system.ID(*filter.ID))
	}
	if filter.Name != nil {
		query = query.Where(system.Name(*filter.Name))
	}
	if filter.Status != nil {
		query = query.Where(system.StatusEQ(system.Status(*filter.Status)))
	}
	return query
}

func fromEntSystem(e *ent.System) *System {
	return &System{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		Status:      Status(e.Status),
		CreatedAt:   e.CreateTime,
		UpdatedAt:   e.UpdateTime,
	}
}

func fromEntSystems(systems []*ent.System) []*System {
	result := make([]*System, len(systems))
	for i, e := range systems {
		result[i] = fromEntSystem(e)
	}
	return result
}
