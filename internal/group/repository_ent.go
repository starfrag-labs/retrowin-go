package group

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	entgroup "github.com/starfrag-lab/retrowin-go/ent/group"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Group, error) {
	entGroup, err := client.Group.Create().
		SetSystemID(cmd.SystemID).
		SetGid(cmd.GID).
		SetGroupname(cmd.Groupname).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}
	return fromEntGroup(entGroup), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id int64) (*Group, error) {
	entGroup, err := client.Group.Query().
		Where(entgroup.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return fromEntGroup(entGroup), nil
}

func (r *EntRepository) GetBySystemIDAndGID(ctx context.Context, client *ent.Client, systemID int64, gid string) (*Group, error) {
	entGroup, err := client.Group.Query().
		Where(
			entgroup.SystemIDEQ(systemID),
			entgroup.Gid(gid),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return fromEntGroup(entGroup), nil
}

func (r *EntRepository) GetBySystemIDAndGroupname(ctx context.Context, client *ent.Client, systemID int64, groupname string) (*Group, error) {
	entGroup, err := client.Group.Query().
		Where(
			entgroup.SystemIDEQ(systemID),
			entgroup.Groupname(groupname),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return fromEntGroup(entGroup), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error {
	builder := client.Group.UpdateOneID(cmd.ID)

	if cmd.Groupname != nil {
		builder.SetGroupname(*cmd.Groupname)
	}

	return builder.Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.Group.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter Filter) ([]*Group, error) {
	query := client.Group.Query()
	query = applyFilter(query, filter)

	entGroups, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find groups: %w", err)
	}
	return fromEntGroups(entGroups), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter Filter) (*Group, error) {
	query := client.Group.Query()
	query = applyFilter(query, filter)

	entGroup, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find group: %w", err)
	}
	return fromEntGroup(entGroup), nil
}

func (r *EntRepository) Exists(ctx context.Context, client *ent.Client, filter Filter) (bool, error) {
	query := client.Group.Query()
	query = applyFilter(query, filter)
	return query.Exist(ctx)
}

func applyFilter(query *ent.GroupQuery, filter Filter) *ent.GroupQuery {
	if filter.ID != nil {
		query = query.Where(entgroup.ID(*filter.ID))
	}
	if filter.SystemID != nil {
		query = query.Where(entgroup.SystemIDEQ(*filter.SystemID))
	}
	if filter.GID != nil {
		query = query.Where(entgroup.Gid(*filter.GID))
	}
	if filter.Groupname != nil {
		query = query.Where(entgroup.Groupname(*filter.Groupname))
	}
	return query
}

func fromEntGroup(e *ent.Group) *Group {
	return &Group{
		ID:        e.ID,
		SystemID:  e.SystemID,
		GID:       e.Gid,
		Groupname: e.Groupname,
		CreatedAt: e.CreateTime,
		UpdatedAt: e.UpdateTime,
	}
}

func fromEntGroups(groups []*ent.Group) []*Group {
	result := make([]*Group, len(groups))
	for i, e := range groups {
		result[i] = fromEntGroup(e)
	}
	return result
}
