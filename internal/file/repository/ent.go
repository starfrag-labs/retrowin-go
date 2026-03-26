package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/starfrag-lab/retrowin-go/ent"
	entfile "github.com/starfrag-lab/retrowin-go/ent/file"
	"github.com/starfrag-lab/retrowin-go/ent/fileinfo"
	"github.com/starfrag-lab/retrowin-go/ent/filepath"
	"github.com/starfrag-lab/retrowin-go/ent/filerole"
	"github.com/starfrag-lab/retrowin-go/internal/file"
)

// EntRepository implements the file.Repository interface using Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a new EntRepository.
func NewEntRepository(client *ent.Client) file.Repository {
	return &EntRepository{client: client}
}

// Create creates a new file.
func (r *EntRepository) Create(ctx context.Context, cmd *file.CreateCommand) (*file.File, error) {
	// Generate file key
	fileKey := uuid.New().String()

	// Build create query
	builder := r.client.File.Create().
		SetFileKey(fileKey).
		SetType(entfile.Type(cmd.Type)).
		SetFileName(cmd.FileName).
		SetOwnerID(cmd.OwnerID).
		SetByteSize(0).
		SetIsSystem(false)

	// Set parent ID if provided
	if cmd.ParentKey != nil && *cmd.ParentKey != "" {
		parent, err := r.client.File.Query().
			Where(entfile.FileKeyEQ(*cmd.ParentKey)).
			Only(ctx)
		if err != nil {
			return nil, err
		}
		if parent != nil {
			builder = builder.SetParentID(int64(parent.ID))
		}
	}

	// Create file
	entFile, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.toFile(entFile), nil
}

// GetByID retrieves a file by ID.
func (r *EntRepository) GetByID(ctx context.Context, id int64) (*file.File, error) {
	entFile, err := r.client.File.Query().
		Where(entfile.IDEQ(int(id))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if entFile == nil {
		return nil, nil
	}

	return r.toFile(entFile), nil
}

// GetByKey retrieves a file by file key.
func (r *EntRepository) GetByKey(ctx context.Context, fileKey string) (*file.File, error) {
	entFile, err := r.client.File.Query().
		Where(entfile.FileKeyEQ(fileKey)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if entFile == nil {
		return nil, nil
	}

	return r.toFile(entFile), nil
}

// GetByOwnerAndSystemType retrieves a system file by owner and type.
func (r *EntRepository) GetByOwnerAndSystemType(ctx context.Context, ownerID int64, systemType string) (*file.File, error) {
	entFile, err := r.client.File.Query().
		Where(
			entfile.OwnerIDEQ(ownerID),
			entfile.IsSystemEQ(true),
			entfile.SystemTypeEQ(systemType),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if entFile == nil {
		return nil, nil
	}

	return r.toFile(entFile), nil
}

// GetChildren retrieves all children of a parent file.
func (r *EntRepository) GetChildren(ctx context.Context, parentID int64) ([]*file.File, error) {
	entFiles, err := r.client.File.Query().
		Where(entfile.ParentIDEQ(parentID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	files := make([]*file.File, len(entFiles))
	for i, ef := range entFiles {
		files[i] = r.toFile(ef)
	}
	return files, nil
}

// Update updates a file.
func (r *EntRepository) Update(ctx context.Context, id int64, cmd *file.UpdateCommand) (*file.File, error) {
	update := r.client.File.UpdateOneID(int(id))

	if cmd.FileName != nil {
		update = update.SetFileName(*cmd.FileName)
	}

	if cmd.ByteSize != nil {
		update = update.SetByteSize(*cmd.ByteSize)
	}

	entFile, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.toFile(entFile), nil
}

// Delete deletes a file by ID.
func (r *EntRepository) Delete(ctx context.Context, id int64) error {
	return r.client.File.DeleteOneID(int(id)).Exec(ctx)
}

// ExistsByKey checks if a file exists by key.
func (r *EntRepository) ExistsByKey(ctx context.Context, fileKey string) (bool, error) {
	return r.client.File.Query().
		Where(entfile.FileKeyEQ(fileKey)).
		Exist(ctx)
}

// GetByOwnerAndParent retrieves files by owner and parent.
func (r *EntRepository) GetByOwnerAndParent(ctx context.Context, ownerID int64, parentID *int64) ([]*file.File, error) {
	query := r.client.File.Query().Where(entfile.OwnerIDEQ(ownerID))

	if parentID == nil {
		query = query.Where(entfile.ParentIDIsNil())
	} else {
		query = query.Where(entfile.ParentIDEQ(*parentID))
	}

	entFiles, err := query.All(ctx)
	if err != nil {
		return nil, err
	}

	files := make([]*file.File, len(entFiles))
	for i, ef := range entFiles {
		files[i] = r.toFile(ef)
	}
	return files, nil
}

// UpdateByteSize updates the byte size of a file.
func (r *EntRepository) UpdateByteSize(ctx context.Context, id int64, byteSize int64) error {
	_, err := r.client.File.UpdateOneID(int(id)).
		SetByteSize(byteSize).
		Save(ctx)
	return err
}

// Move moves a file to a new parent.
func (r *EntRepository) Move(ctx context.Context, fileID int64, newParentID int64) error {
	_, err := r.client.File.UpdateOneID(int(fileID)).
		SetParentID(newParentID).
		Save(ctx)
	return err
}

// Copy copies a file to a new parent.
func (r *EntRepository) Copy(ctx context.Context, fileID int64, newParentID int64, ownerID int64) (*file.File, error) {
	// Get original file
	original, err := r.client.File.Query().
		Where(entfile.IDEQ(int(fileID))).
		Only(ctx)
	if err != nil {
		return nil, err
	}

	// Create new file with copied data
	newFileKey := uuid.New().String()
	newFile, err := r.client.File.Create().
		SetFileKey(newFileKey).
		SetType(original.Type).
		SetFileName(original.FileName).
		SetParentID(newParentID).
		SetOwnerID(ownerID).
		SetByteSize(original.ByteSize).
		SetIsSystem(false).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return r.toFile(newFile), nil
}

// toFile converts an Ent file to a domain file.
func (r *EntRepository) toFile(ef *ent.File) *file.File {
	var parentID *int64
	if ef.ParentID != nil {
		parentID = ef.ParentID
	}

	var systemType *string
	if ef.SystemType != nil {
		systemType = ef.SystemType
	}

	return &file.File{
		ID:         int64(ef.ID),
		FileKey:    ef.FileKey,
		Type:       file.FileType(ef.Type),
		FileName:   ef.FileName,
		OwnerID:    ef.OwnerID,
		ParentID:   parentID,
		ByteSize:   ef.ByteSize,
		IsSystem:   ef.IsSystem,
		SystemType: systemType,
		CreatedAt:  ef.CreateTime,
		UpdatedAt:  ef.UpdateTime,
	}
}

// EntFileInfoRepository implements the file.FileInfoRepository interface.
type EntFileInfoRepository struct {
	client *ent.Client
}

// NewEntFileInfoRepository creates a new EntFileInfoRepository.
func NewEntFileInfoRepository(client *ent.Client) file.FileInfoRepository {
	return &EntFileInfoRepository{client: client}
}

// Create creates file info.
func (r *EntFileInfoRepository) Create(ctx context.Context, fileID int64, byteSize int64) (*file.FileInfo, error) {
	info, err := r.client.FileInfo.Create().
		SetFileID(fileID).
		SetByteSize(byteSize).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return &file.FileInfo{
		FileID:     info.FileID,
		CreateDate: info.CreateDate,
		UpdateDate: info.UpdateDate,
		ByteSize:   info.ByteSize,
	}, nil
}

// GetByFileID retrieves file info by file ID.
func (r *EntFileInfoRepository) GetByFileID(ctx context.Context, fileID int64) (*file.FileInfo, error) {
	info, err := r.client.FileInfo.Query().
		Where(fileinfo.FileIDEQ(fileID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if info == nil {
		return nil, nil
	}

	return &file.FileInfo{
		FileID:     info.FileID,
		CreateDate: info.CreateDate,
		UpdateDate: info.UpdateDate,
		ByteSize:   info.ByteSize,
	}, nil
}

// Update updates file info.
func (r *EntFileInfoRepository) Update(ctx context.Context, fileID int64, byteSize int64) (*file.FileInfo, error) {
	info, err := r.client.FileInfo.UpdateOneID(int(fileID)).
		SetByteSize(byteSize).
		SetUpdateDate(time.Now()).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	return &file.FileInfo{
		FileID:     info.FileID,
		CreateDate: info.CreateDate,
		UpdateDate: info.UpdateDate,
		ByteSize:   info.ByteSize,
	}, nil
}

// Delete deletes file info.
func (r *EntFileInfoRepository) Delete(ctx context.Context, fileID int64) error {
	return r.client.FileInfo.DeleteOneID(int(fileID)).Exec(ctx)
}

// EntFilePathRepository implements the file.FilePathRepository interface.
type EntFilePathRepository struct {
	client *ent.Client
}

// NewEntFilePathRepository creates a new EntFilePathRepository.
func NewEntFilePathRepository(client *ent.Client) file.FilePathRepository {
	return &EntFilePathRepository{client: client}
}

// Create creates a file path.
func (r *EntFilePathRepository) Create(ctx context.Context, fileID int64, path []int64) error {
	_, err := r.client.FilePath.Create().
		SetFileID(fileID).
		SetPath(path).
		Save(ctx)
	if err != nil {
		// If already exists, update
		if ent.IsConstraintError(err) {
			return r.Update(ctx, fileID, path)
		}
		return err
	}
	return nil
}

// GetByFileID retrieves the file path by file ID.
func (r *EntFilePathRepository) GetByFileID(ctx context.Context, fileID int64) ([]int64, error) {
	fp, err := r.client.FilePath.Query().
		Where(filepath.FileIDEQ(fileID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if fp == nil {
		return nil, nil
	}

	return fp.Path, nil
}

// Update updates a file path.
func (r *EntFilePathRepository) Update(ctx context.Context, fileID int64, path []int64) error {
	_, err := r.client.FilePath.Update().
		Where(filepath.FileIDEQ(fileID)).
		SetPath(path).
		Save(ctx)
	return err
}

// Delete deletes a file path.
func (r *EntFilePathRepository) Delete(ctx context.Context, fileID int64) error {
	_, err := r.client.FilePath.Delete().
		Where(filepath.FileIDEQ(fileID)).
		Exec(ctx)
	return err
}

// EntFileRoleRepository implements the file.FileRoleRepository interface.
type EntFileRoleRepository struct {
	client *ent.Client
}

// NewEntFileRoleRepository creates a new EntFileRoleRepository.
func NewEntFileRoleRepository(client *ent.Client) file.FileRoleRepository {
	return &EntFileRoleRepository{client: client}
}

// Create creates file roles.
func (r *EntFileRoleRepository) Create(ctx context.Context, userID int64, fileID int64, roles []string) error {
	_, err := r.client.FileRole.Create().
		SetUserID(userID).
		SetFileID(fileID).
		SetRoles(roles).
		Save(ctx)
	if err != nil {
		// If already exists, update
		if ent.IsConstraintError(err) {
			return r.Update(ctx, userID, fileID, roles)
		}
	}
	return err
}

// GetByUserAndFile retrieves roles for a user and file.
func (r *EntFileRoleRepository) GetByUserAndFile(ctx context.Context, userID int64, fileID int64) ([]string, error) {
	fr, err := r.client.FileRole.Query().
		Where(
			filerole.UserIDEQ(userID),
			filerole.FileIDEQ(fileID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if fr == nil {
		return nil, nil
	}

	return fr.Roles, nil
}

// Update updates file roles.
func (r *EntFileRoleRepository) Update(ctx context.Context, userID int64, fileID int64, roles []string) error {
	_, err := r.client.FileRole.Update().
		Where(
			filerole.UserIDEQ(userID),
			filerole.FileIDEQ(fileID),
		).
		SetRoles(roles).
		Save(ctx)
	return err
}

// Delete deletes file roles.
func (r *EntFileRoleRepository) Delete(ctx context.Context, userID int64, fileID int64) error {
	_, err := r.client.FileRole.Delete().
		Where(
			filerole.UserIDEQ(userID),
			filerole.FileIDEQ(fileID),
		).
		Exec(ctx)
	return err
}

// DeleteByFile deletes all roles for a file.
func (r *EntFileRoleRepository) DeleteByFile(ctx context.Context, fileID int64) error {
	_, err := r.client.FileRole.Delete().
		Where(filerole.FileIDEQ(fileID)).
		Exec(ctx)
	return err
}
