package file

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/file"
	"github.com/starfrag-lab/retrowin-go/ent/fileinfo"
	"github.com/starfrag-lab/retrowin-go/ent/filepath"
	"github.com/starfrag-lab/retrowin-go/ent/filerole"
)

// EntRepository implements Repository using Ent.
type EntRepository struct {
	client *ent.Client
}

// NewEntRepository creates a new EntRepository.
func NewEntRepository(client *ent.Client) Repository {
	return &EntRepository{client: client}
}

// Create creates a new file.
func (r *EntRepository) Create(ctx context.Context, cmd *CreateCommand) (*File, error) {
	fileKey := uuid.New().String()

	builder := r.client.File.Create().
		SetFileKey(fileKey).
		SetType(file.Type(cmd.Type)).
		SetFileName(cmd.FileName).
		SetOwnerID(cmd.OwnerID).
		SetIsSystem(false)

	if cmd.ParentKey != nil {
		parent, err := r.client.File.Query().
			Where(file.FileKey(*cmd.ParentKey)).
			Only(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to find parent: %w", err)
		}
		parentID := int64(parent.ID)
		builder.SetParentID(parentID)
	}

	entFile, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return fromEntFile(entFile), nil
}

// GetByID retrieves a file by ID.
func (r *EntRepository) GetByID(ctx context.Context, id int64) (*File, error) {
	entFile, err := r.client.File.Query().
		Where(file.ID(int(id))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return fromEntFile(entFile), nil
}

// GetByKey retrieves a file by file key.
func (r *EntRepository) GetByKey(ctx context.Context, fileKey string) (*File, error) {
	entFile, err := r.client.File.Query().
		Where(file.FileKey(fileKey)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return fromEntFile(entFile), nil
}

// GetByOwnerAndSystemType retrieves a system file by owner and type.
func (r *EntRepository) GetByOwnerAndSystemType(ctx context.Context, ownerID int64, systemType string) (*File, error) {
	entFile, err := r.client.File.Query().
		Where(
			file.OwnerIDEQ(ownerID),
			file.IsSystem(true),
			file.SystemType(systemType),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return fromEntFile(entFile), nil
}

// GetChildren retrieves all children of a container.
func (r *EntRepository) GetChildren(ctx context.Context, parentID int64) ([]*File, error) {
	entFiles, err := r.client.File.Query().
		Where(file.ParentIDEQ(parentID)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	return fromEntFiles(entFiles), nil
}

// Update updates a file.
func (r *EntRepository) Update(ctx context.Context, id int64, cmd *UpdateCommand) (*File, error) {
	builder := r.client.File.UpdateOneID(int(id))

	if cmd.FileName != nil {
		builder.SetFileName(*cmd.FileName)
	}
	if cmd.ByteSize != nil {
		builder.SetByteSize(*cmd.ByteSize)
	}

	entFile, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update file: %w", err)
	}
	return fromEntFile(entFile), nil
}

// Delete deletes a file by ID.
func (r *EntRepository) Delete(ctx context.Context, id int64) error {
	return r.client.File.DeleteOneID(int(id)).Exec(ctx)
}

// ExistsByKey checks if a file exists by key.
func (r *EntRepository) ExistsByKey(ctx context.Context, fileKey string) (bool, error) {
	return r.client.File.Query().
		Where(file.FileKey(fileKey)).
		Exist(ctx)
}

// GetByOwnerAndParent retrieves files by owner and parent.
func (r *EntRepository) GetByOwnerAndParent(ctx context.Context, ownerID int64, parentID *int64) ([]*File, error) {
	query := r.client.File.Query().Where(file.OwnerIDEQ(ownerID))

	if parentID == nil {
		query = query.Where(file.ParentIDIsNil())
	} else {
		query = query.Where(file.ParentIDEQ(*parentID))
	}

	entFiles, err := query.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}
	return fromEntFiles(entFiles), nil
}

// UpdateByteSize updates the byte size of a file.
func (r *EntRepository) UpdateByteSize(ctx context.Context, id int64, byteSize int64) error {
	return r.client.File.UpdateOneID(int(id)).
		SetByteSize(byteSize).
		Exec(ctx)
}

// Move moves a file to a new parent.
func (r *EntRepository) Move(ctx context.Context, fileID int64, newParentID int64) error {
	return r.client.File.UpdateOneID(int(fileID)).
		SetParentID(newParentID).
		Exec(ctx)
}

// Copy copies a file to a new parent and returns the new file.
func (r *EntRepository) Copy(ctx context.Context, fileID int64, newParentID int64, ownerID int64) (*File, error) {
	// Get original file
	original, err := r.client.File.Query().
		Where(file.ID(int(fileID))).
		Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get original file: %w", err)
	}

	// Create copy
	newFileKey := uuid.New().String()
	copy, err := r.client.File.Create().
		SetFileKey(newFileKey).
		SetType(original.Type).
		SetFileName(original.FileName).
		SetOwnerID(ownerID).
		SetParentID(newParentID).
		SetByteSize(original.ByteSize).
		SetIsSystem(false).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	return fromEntFile(copy), nil
}

func fromEntFile(e *ent.File) *File {
	return &File{
		ID:         int64(e.ID),
		FileKey:    e.FileKey,
		Type:       FileType(e.Type),
		FileName:   e.FileName,
		OwnerID:    e.OwnerID,
		ParentID:   e.ParentID,
		ByteSize:   e.ByteSize,
		IsSystem:   e.IsSystem,
		SystemType: e.SystemType,
		CreatedAt:  e.CreateTime,
		UpdatedAt:  e.UpdateTime,
	}
}

func fromEntFiles(files []*ent.File) []*File {
	result := make([]*File, len(files))
	for i, f := range files {
		result[i] = fromEntFile(f)
	}
	return result
}

// EntFileInfoRepository implements FileInfoRepository using Ent.
type EntFileInfoRepository struct {
	client *ent.Client
}

// NewEntFileInfoRepository creates a new EntFileInfoRepository.
func NewEntFileInfoRepository(client *ent.Client) FileInfoRepository {
	return &EntFileInfoRepository{client: client}
}

// Create creates file info for a file.
func (r *EntFileInfoRepository) Create(ctx context.Context, fileID int64, byteSize int64) (*FileInfo, error) {
	info, err := r.client.FileInfo.Create().
		SetFileID(fileID).
		SetByteSize(byteSize).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create file info: %w", err)
	}
	return fromEntFileInfo(info), nil
}

// GetByFileID retrieves file info by file ID.
func (r *EntFileInfoRepository) GetByFileID(ctx context.Context, fileID int64) (*FileInfo, error) {
	info, err := r.client.FileInfo.Query().
		Where(fileinfo.FileIDEQ(fileID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	return fromEntFileInfo(info), nil
}

// Update updates file info.
func (r *EntFileInfoRepository) Update(ctx context.Context, fileID int64, byteSize int64) (*FileInfo, error) {
	err := r.client.FileInfo.Update().
		Where(fileinfo.FileIDEQ(fileID)).
		SetByteSize(byteSize).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to update file info: %w", err)
	}
	// Fetch the updated record
	return r.GetByFileID(ctx, fileID)
}

// Delete deletes file info by file ID.
func (r *EntFileInfoRepository) Delete(ctx context.Context, fileID int64) error {
	_, err := r.client.FileInfo.Delete().
		Where(fileinfo.FileIDEQ(fileID)).
		Exec(ctx)
	return err
}

func fromEntFileInfo(e *ent.FileInfo) *FileInfo {
	return &FileInfo{
		FileID:     e.FileID,
		CreateDate: e.CreateDate,
		UpdateDate: e.UpdateDate,
		ByteSize:   e.ByteSize,
	}
}

// EntFilePathRepository implements FilePathRepository using Ent.
type EntFilePathRepository struct {
	client *ent.Client
}

// NewEntFilePathRepository creates a new EntFilePathRepository.
func NewEntFilePathRepository(client *ent.Client) FilePathRepository {
	return &EntFilePathRepository{client: client}
}

// Create creates a file path for a file.
func (r *EntFilePathRepository) Create(ctx context.Context, fileID int64, path []int64) error {
	return r.client.FilePath.Create().
		SetFileID(fileID).
		SetPath(path).
		Exec(ctx)
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
		return nil, fmt.Errorf("failed to get file path: %w", err)
	}
	return fp.Path, nil
}

// Update updates file path.
func (r *EntFilePathRepository) Update(ctx context.Context, fileID int64, path []int64) error {
	return r.client.FilePath.Update().
		Where(filepath.FileIDEQ(fileID)).
		SetPath(path).
		Exec(ctx)
}

// Delete deletes file path by file ID.
func (r *EntFilePathRepository) Delete(ctx context.Context, fileID int64) error {
	_, err := r.client.FilePath.Delete().
		Where(filepath.FileIDEQ(fileID)).
		Exec(ctx)
	return err
}

// EntFileRoleRepository implements FileRoleRepository using Ent.
type EntFileRoleRepository struct {
	client *ent.Client
}

// NewEntFileRoleRepository creates a new EntFileRoleRepository.
func NewEntFileRoleRepository(client *ent.Client) FileRoleRepository {
	return &EntFileRoleRepository{client: client}
}

// Create creates a file role.
func (r *EntFileRoleRepository) Create(ctx context.Context, userID int64, fileID int64, roles []string) error {
	return r.client.FileRole.Create().
		SetUserID(userID).
		SetFileID(fileID).
		SetRoles(roles).
		Exec(ctx)
}

// GetByUserAndFile retrieves roles for a user on a file.
func (r *EntFileRoleRepository) GetByUserAndFile(ctx context.Context, userID int64, fileID int64) ([]string, error) {
	role, err := r.client.FileRole.Query().
		Where(
			filerole.UserIDEQ(userID),
			filerole.FileIDEQ(fileID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file role: %w", err)
	}
	return role.Roles, nil
}

// Update updates file roles.
func (r *EntFileRoleRepository) Update(ctx context.Context, userID int64, fileID int64, roles []string) error {
	return r.client.FileRole.Update().
		Where(
			filerole.UserIDEQ(userID),
			filerole.FileIDEQ(fileID),
		).
		SetRoles(roles).
		Exec(ctx)
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
