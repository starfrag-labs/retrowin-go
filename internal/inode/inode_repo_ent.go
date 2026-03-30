package inode

import (
	"context"
	"fmt"
	"time"

	"github.com/starfrag-lab/retrowin-go/ent"
	entinode "github.com/starfrag-lab/retrowin-go/ent/inode"
)

// timeNow returns the current time. Extracted for testability.
var timeNow = time.Now

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Inode, error) {
	now := timeNow()

	builder := client.Inode.Create().
		SetSystemID(params.SystemID).
		SetMode(params.Mode).
		SetUID(params.UID).
		SetGid(params.GID).
		SetSize(0).
		SetLinkCount(1).
		SetFlags(params.Flags).
		SetAtime(now).
		SetMtime(now).
		SetCtime(now)

	if params.Content != nil {
		builder.SetContent(params.Content)
	}

	entInode, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create inode: %w", err)
	}

	return fromEnt(entInode), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id int64) (*Inode, error) {
	entInode, err := client.Inode.Query().
		Where(entinode.ID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get inode: %w", err)
	}
	return fromEnt(entInode), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, params *UpdateParams) error {
	builder := client.Inode.UpdateOneID(params.ID)

	if params.Mode != nil {
		builder.SetMode(*params.Mode)
	}
	if params.UID != nil {
		builder.SetUID(*params.UID)
	}
	if params.GID != nil {
		builder.SetGid(*params.GID)
	}
	if params.Size != nil {
		builder.SetSize(*params.Size)
	}
	if params.Flags != nil {
		builder.SetFlags(*params.Flags)
	}
	if params.Atime != nil {
		builder.SetAtime(*params.Atime)
	}
	if params.Mtime != nil {
		builder.SetMtime(*params.Mtime)
	}
	if params.Ctime != nil {
		builder.SetCtime(*params.Ctime)
	}

	return builder.Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.Inode.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter *QueryFilter) ([]*Inode, error) {
	query := client.Inode.Query()
	query = applyFilter(query, filter)

	entInodes, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find inodes: %w", err)
	}
	return fromEntSlice(entInodes), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter *QueryFilter) (*Inode, error) {
	query := client.Inode.Query()
	query = applyFilter(query, filter)

	entInode, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find inode: %w", err)
	}
	return fromEnt(entInode), nil
}

func (r *EntRepository) UpdateLinkCount(ctx context.Context, client *ent.Client, id int64, delta int) error {
	return client.Inode.UpdateOneID(id).
		AddLinkCount(delta).
		Exec(ctx)
}

func applyFilter(query *ent.InodeQuery, filter *QueryFilter) *ent.InodeQuery {
	if filter == nil {
		return query
	}
	if filter.ID != nil {
		query = query.Where(entinode.ID(*filter.ID))
	}
	if filter.SystemID != nil {
		query = query.Where(entinode.SystemIDEQ(*filter.SystemID))
	}
	if filter.UID != nil {
		query = query.Where(entinode.UIDEQ(*filter.UID))
	}
	if filter.GID != nil {
		query = query.Where(entinode.GidEQ(*filter.GID))
	}
	return query
}

func fromEnt(e *ent.Inode) *Inode {
	var content []byte
	if e.Content != nil {
		content = e.Content
	}

	return NewInode(
		e.ID,
		e.SystemID,
		e.Mode,
		e.UID,
		e.Gid,
		e.Size,
		e.LinkCount,
		e.Flags,
		e.Atime,
		e.Mtime,
		e.Ctime,
		content,
		e.CreateTime,
		e.UpdateTime,
	)
}

func fromEntSlice(inodes []*ent.Inode) []*Inode {
	result := make([]*Inode, len(inodes))
	for i, e := range inodes {
		result[i] = fromEnt(e)
	}
	return result
}
