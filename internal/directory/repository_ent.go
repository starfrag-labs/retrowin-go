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

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Entry, error) {
	entEntry, err := client.DirectoryEntry.Create().
		SetParentID(cmd.ParentID).
		SetName(cmd.Name).
		SetChildID(cmd.ChildID).
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

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, entry *Entry) error {
	return client.DirectoryEntry.UpdateOneID(entry.ID).
		SetParentID(entry.ParentID).
		SetName(entry.Name).
		SetChildID(entry.ChildID).
		Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.DirectoryEntry.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter Filter) ([]*Entry, error) {
	query := client.DirectoryEntry.Query()
	query = applyFilter(query, filter)

	entEntries, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find directory entries: %w", err)
	}
	return fromEntSlice(entEntries), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter Filter) (*Entry, error) {
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

func (r *EntRepository) Exists(ctx context.Context, client *ent.Client, filter Filter) (bool, error) {
	query := client.DirectoryEntry.Query()
	query = applyFilter(query, filter)
	return query.Exist(ctx)
}

func applyFilter(query *ent.DirectoryEntryQuery, filter Filter) *ent.DirectoryEntryQuery {
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
	return &Entry{
		ID:       e.ID,
		ParentID: e.ParentID,
		Name:     e.Name,
		ChildID:  e.ChildID,
	}
}

func fromEntSlice(entries []*ent.DirectoryEntry) []*Entry {
	result := make([]*Entry, len(entries))
	for i, e := range entries {
		result[i] = fromEnt(e)
	}
	return result
}
