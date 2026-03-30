package directory

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/directoryentry"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Entry, error) {
	entEntry, err := client.DirectoryEntry.Create().
		SetParentID(params.ParentID).
		SetName(params.Name).
		SetChildID(params.ChildID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory entry: %w", err)
	}
	return fromEnt(entEntry), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id int64) (*Entry, error) {
	entEntry, err := client.DirectoryEntry.Query().
		Where(directoryentry.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get directory entry: %w", err)
	}
	return fromEnt(entEntry), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, params *UpdateParams) error {
	builder := client.DirectoryEntry.UpdateOneID(params.ID)

	if params.ParentID != nil {
		builder.SetParentID(*params.ParentID)
	}
	if params.Name != nil {
		builder.SetName(*params.Name)
	}
	if params.ChildID != nil {
		builder.SetChildID(*params.ChildID)
	}

	return builder.Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.DirectoryEntry.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Entry, error) {
	query := client.DirectoryEntry.Query()
	query = applyFilter(query, filter)

	entEntries, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find directory entries: %w", err)
	}
	return fromEntSlice(entEntries), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Entry, error) {
	query := client.DirectoryEntry.Query()
	query = applyFilter(query, filter)

	entEntry, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find directory entry: %w", err)
	}
	return fromEnt(entEntry), nil
}

func (r *EntRepository) Exists(ctx context.Context, client *ent.Client, filter *QueryFilter) (bool, error) {
	query := client.DirectoryEntry.Query()
	query = applyFilter(query, filter)
	return query.Exist(ctx)
}

func applyFilter(query *ent.DirectoryEntryQuery, filter *QueryFilter) *ent.DirectoryEntryQuery {
	if filter == nil {
		return query
	}
	if filter.ID != nil {
		query = query.Where(directoryentry.ID(*filter.ID))
	}
	if filter.ParentID != nil {
		query = query.Where(directoryentry.ParentIDEQ(*filter.ParentID))
	}
	if filter.Name != nil {
		query = query.Where(directoryentry.Name(*filter.Name))
	}
	if filter.ChildID != nil {
		query = query.Where(directoryentry.ChildIDEQ(*filter.ChildID))
	}
	return query
}

func fromEnt(e *ent.DirectoryEntry) *Entry {
	return NewEntry(
		e.ID,
		e.ParentID,
		e.Name,
		e.ChildID,
	)
}

func fromEntSlice(entries []*ent.DirectoryEntry) []*Entry {
	result := make([]*Entry, len(entries))
	for i, e := range entries {
		result[i] = fromEnt(e)
	}
	return result
}
