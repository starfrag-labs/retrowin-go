package inode

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/inode"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *CreateParams) (*Inode, error) {
	builder := client.Inode.Create().
		SetFileType(inode.FileType(params.FileType)).
		SetOwnerUID(params.OwnerUID).
		SetOwnerGid(params.OwnerGID).
		SetPermOwner(params.PermOwner).
		SetPermGroup(params.PermGroup).
		SetPermOthers(params.PermOthers).
		SetByteSize(0).
		SetLinkCount(1).
		SetIsSystem(params.IsSystem)

	if params.SystemID != nil {
		builder.SetSystemID(*params.SystemID)
	}
	if params.SystemType != nil {
		builder.SetSystemType(*params.SystemType)
	}

	entInode, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create inode: %w", err)
	}

	return fromEnt(entInode), nil
}

func (r *EntRepository) GetByID(ctx context.Context, client *ent.Client, id int64) (*Inode, error) {
	entInode, err := client.Inode.Query().
		Where(inode.ID(id)).
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

	if params.ByteSize != nil {
		builder.SetByteSize(*params.ByteSize)
	}
	if params.PermOwner != nil {
		builder.SetPermOwner(*params.PermOwner)
	}
	if params.PermGroup != nil {
		builder.SetPermGroup(*params.PermGroup)
	}
	if params.PermOthers != nil {
		builder.SetPermOthers(*params.PermOthers)
	}
	if params.LinkCount != nil {
		builder.SetLinkCount(*params.LinkCount)
	}
	if params.AccessedAt != nil {
		builder.SetAccessedAt(*params.AccessedAt)
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

func (r *EntRepository) UpdateLinkCount(ctx context.Context, client *ent.Client, id int64, delta int16) error {
	return client.Inode.UpdateOneID(id).
		AddLinkCount(delta).
		Exec(ctx)
}

func applyFilter(query *ent.InodeQuery, filter *QueryFilter) *ent.InodeQuery {
	if filter == nil {
		return query
	}
	if filter.ID != nil {
		query = query.Where(inode.ID(*filter.ID))
	}
	if filter.SystemID != nil {
		query = query.Where(inode.SystemIDEQ(*filter.SystemID))
	}
	if filter.OwnerUID != nil {
		query = query.Where(inode.OwnerUID(*filter.OwnerUID))
	}
	if filter.IsSystem != nil {
		query = query.Where(inode.IsSystem(*filter.IsSystem))
	}
	if filter.SystemType != nil {
		query = query.Where(inode.SystemType(*filter.SystemType))
	}
	if filter.FileType != nil {
		query = query.Where(inode.FileTypeEQ(inode.FileType(string(*filter.FileType))))
	}
	return query
}

func fromEnt(e *ent.Inode) *Inode {
	return NewInode(
		e.ID,
		e.SystemID,
		FileType(e.FileType),
		e.ByteSize,
		e.OwnerUID,
		e.OwnerGid,
		e.PermOwner,
		e.PermGroup,
		e.PermOthers,
		e.LinkCount,
		e.AccessedAt,
		e.IsSystem,
		e.SystemType,
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
