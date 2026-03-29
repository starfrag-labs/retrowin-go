package inode

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/filedata"
	"github.com/starfrag-lab/retrowin-go/ent/inode"
)

// ==================== Inode Repository ====================

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, cmd *CreateCommand) (*Inode, error) {
	builder := client.Inode.Create().
		SetFileType(inode.FileType(cmd.FileType)).
		SetOwnerUID(cmd.OwnerUID).
		SetOwnerGid(cmd.OwnerGID).
		SetPermOwner(cmd.PermOwner).
		SetPermGroup(cmd.PermGroup).
		SetPermOthers(cmd.PermOthers).
		SetByteSize(0).
		SetLinkCount(1).
		SetIsSystem(cmd.IsSystem)

	if cmd.SystemID != nil {
		builder.SetSystemID(*cmd.SystemID)
	}
	if cmd.SystemType != nil {
		builder.SetSystemType(*cmd.SystemType)
	}

	entInode, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create inode: %w", err)
	}

	return fromEntInode(entInode), nil
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
	return fromEntInode(entInode), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, cmd *UpdateCommand) error {
	builder := client.Inode.UpdateOneID(cmd.ID)

	if cmd.ByteSize != nil {
		builder.SetByteSize(*cmd.ByteSize)
	}
	if cmd.PermOwner != nil {
		builder.SetPermOwner(*cmd.PermOwner)
	}
	if cmd.PermGroup != nil {
		builder.SetPermGroup(*cmd.PermGroup)
	}
	if cmd.PermOthers != nil {
		builder.SetPermOthers(*cmd.PermOthers)
	}
	if cmd.LinkCount != nil {
		builder.SetLinkCount(*cmd.LinkCount)
	}
	if cmd.AccessedAt != nil {
		builder.SetAccessedAt(*cmd.AccessedAt)
	}

	return builder.Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, id int64) error {
	return client.Inode.DeleteOneID(id).Exec(ctx)
}

func (r *EntRepository) Find(ctx context.Context, client *ent.Client, filter Filter) ([]*Inode, error) {
	query := client.Inode.Query()
	query = applyFilter(query, filter)

	entInodes, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find inodes: %w", err)
	}
	return fromEntInodes(entInodes), nil
}

func (r *EntRepository) FindOne(ctx context.Context, client *ent.Client, filter Filter) (*Inode, error) {
	query := client.Inode.Query()
	query = applyFilter(query, filter)

	entInode, err := query.Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find inode: %w", err)
	}
	return fromEntInode(entInode), nil
}

func (r *EntRepository) UpdateLinkCount(ctx context.Context, client *ent.Client, id int64, delta int16) error {
	return client.Inode.UpdateOneID(id).
		AddLinkCount(delta).
		Exec(ctx)
}

func applyFilter(query *ent.InodeQuery, filter Filter) *ent.InodeQuery {
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

// ==================== FileData Repository ====================

// EntFileDataRepository implements FileDataRepository using Ent.
type EntFileDataRepository struct{}

// NewEntFileDataRepository creates a new EntFileDataRepository.
func NewEntFileDataRepository() FileDataRepository {
	return &EntFileDataRepository{}
}

func (r *EntFileDataRepository) Create(ctx context.Context, client *ent.Client, cmd *CreateFileDataCommand) (*FileData, error) {
	builder := client.FileData.Create().
		SetInodeID(cmd.InodeID).
		SetStorageType(filedata.StorageType(cmd.StorageType)).
		SetLocation(cmd.Location)

	if cmd.Checksum != nil {
		builder.SetChecksum(*cmd.Checksum)
	}

	entFileData, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create file data: %w", err)
	}
	return fromEntFileData(entFileData), nil
}

func (r *EntFileDataRepository) GetByInodeID(ctx context.Context, client *ent.Client, inodeID int64) (*FileData, error) {
	entFileData, err := client.FileData.Query().
		Where(filedata.InodeIDEQ(inodeID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file data: %w", err)
	}
	return fromEntFileData(entFileData), nil
}

func (r *EntFileDataRepository) Update(ctx context.Context, client *ent.Client, cmd *UpdateFileDataCommand) error {
	builder := client.FileData.Update().
		Where(filedata.InodeIDEQ(cmd.InodeID))

	if cmd.StorageType != nil {
		builder.SetStorageType(filedata.StorageType(*cmd.StorageType))
	}
	if cmd.Location != nil {
		builder.SetLocation(*cmd.Location)
	}
	if cmd.Checksum != nil {
		builder.SetChecksum(*cmd.Checksum)
	}

	return builder.Exec(ctx)
}

func (r *EntFileDataRepository) Delete(ctx context.Context, client *ent.Client, inodeID int64) error {
	_, err := client.FileData.Delete().
		Where(filedata.InodeIDEQ(inodeID)).
		Exec(ctx)
	return err
}

// ==================== Converters ====================

func fromEntInode(e *ent.Inode) *Inode {
	return &Inode{
		ID:          e.ID,
		SystemID:    e.SystemID,
		FileType:    FileType(e.FileType),
		ByteSize:    e.ByteSize,
		OwnerUID:    e.OwnerUID,
		OwnerGID:    e.OwnerGid,
		PermOwner:   e.PermOwner,
		PermGroup:   e.PermGroup,
		PermOthers:  e.PermOthers,
		LinkCount:   e.LinkCount,
		AccessedAt:  e.AccessedAt,
		IsSystem:    e.IsSystem,
		SystemType:  e.SystemType,
		CreatedAt:   e.CreateTime,
		UpdatedAt:   e.UpdateTime,
	}
}

func fromEntInodes(inodes []*ent.Inode) []*Inode {
	result := make([]*Inode, len(inodes))
	for i, e := range inodes {
		result[i] = fromEntInode(e)
	}
	return result
}

func fromEntFileData(e *ent.FileData) *FileData {
	return &FileData{
		InodeID:     e.InodeID,
		StorageType: StorageType(e.StorageType),
		Location:    e.Location,
		Checksum:    e.Checksum,
	}
}
