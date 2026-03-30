package filedata

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/filedata"
)

// EntRepository implements Repository using Ent.
type EntRepository struct{}

// NewEntRepository creates a new EntRepository.
func NewEntRepository() Repository {
	return &EntRepository{}
}

func (r *EntRepository) Create(ctx context.Context, client *ent.Client, params *CreateParams) (*FileData, error) {
	builder := client.FileData.Create().
		SetInodeID(params.InodeID).
		SetStorageType(filedata.StorageType(params.StorageType)).
		SetLocation(params.Location)

	if params.Checksum != nil {
		builder.SetChecksum(*params.Checksum)
	}

	entFileData, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create file data: %w", err)
	}
	return fromEnt(entFileData), nil
}

func (r *EntRepository) GetByInodeID(ctx context.Context, client *ent.Client, inodeID int64) (*FileData, error) {
	entFileData, err := client.FileData.Query().
		Where(filedata.InodeIDEQ(inodeID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file data: %w", err)
	}
	return fromEnt(entFileData), nil
}

func (r *EntRepository) Update(ctx context.Context, client *ent.Client, params *UpdateParams) error {
	builder := client.FileData.Update().
		Where(filedata.InodeIDEQ(params.InodeID))

	if params.StorageType != nil {
		builder.SetStorageType(filedata.StorageType(*params.StorageType))
	}
	if params.Location != nil {
		builder.SetLocation(*params.Location)
	}
	if params.Checksum != nil {
		builder.SetChecksum(*params.Checksum)
	}

	return builder.Exec(ctx)
}

func (r *EntRepository) Delete(ctx context.Context, client *ent.Client, inodeID int64) error {
	_, err := client.FileData.Delete().
		Where(filedata.InodeIDEQ(inodeID)).
		Exec(ctx)
	return err
}

func fromEnt(e *ent.FileData) *FileData {
	return NewFileData(
		e.InodeID,
		StorageType(e.StorageType),
		e.Location,
		e.Checksum,
	)
}
