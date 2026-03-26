package file

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Service defines the interface for file operations.
type Service interface {
	// Get retrieves a file by file key.
	Get(ctx context.Context, fileKey string) (*File, error)

	// GetByID retrieves a file by ID.
	GetByID(ctx context.Context, id int64) (*File, error)

	// GetRoot retrieves the root container for a user.
	GetRoot(ctx context.Context, ownerID int64) (*File, error)

	// GetHome retrieves the home container for a user.
	GetHome(ctx context.Context, ownerID int64) (*File, error)

	// GetTrash retrieves the trash container for a user.
	GetTrash(ctx context.Context, ownerID int64) (*File, error)

	// GetChildren retrieves all children of a container.
	GetChildren(ctx context.Context, fileKey string) ([]*File, error)

	// Create creates a new file or container.
	Create(ctx context.Context, cmd *CreateCommand) (*File, error)

	// Update updates a file's metadata.
	Update(ctx context.Context, fileKey string, cmd *UpdateCommand) (*File, error)

	// Delete deletes a file (moves to trash or permanent).
	Delete(ctx context.Context, fileKey string, permanent bool) error

	// Move moves a file to a different container.
	Move(ctx context.Context, fileKey string, cmd *MoveCommand) (*File, error)

	// Copy copies a file to a different container.
	Copy(ctx context.Context, fileKey string, cmd *CopyCommand) (*File, error)
}

// Errors
var (
	ErrFileNotFound       = errors.New("file not found")
	ErrParentNotFound     = errors.New("parent not found")
	ErrNotContainer       = errors.New("file is not a container")
	ErrAccessDenied       = errors.New("access denied")
	ErrTrashNotFound      = errors.New("trash container not found")
	ErrTargetNotFound     = errors.New("target container not found")
	ErrCannotDeleteSystem = errors.New("cannot delete system files")
	ErrCannotMoveIntoSelf = errors.New("cannot move file into itself")
)

type service struct {
	fileRepo Repository
	infoRepo FileInfoRepository
	pathRepo FilePathRepository
	roleRepo FileRoleRepository
}

// NewService creates a new file service.
func NewService(
	fileRepo Repository,
	infoRepo FileInfoRepository,
	pathRepo FilePathRepository,
	roleRepo FileRoleRepository,
) Service {
	return &service{
		fileRepo: fileRepo,
		infoRepo: infoRepo,
		pathRepo: pathRepo,
		roleRepo: roleRepo,
	}
}

func (s *service) Get(ctx context.Context, fileKey string) (*File, error) {
	file, err := s.fileRepo.GetByKey(ctx, fileKey)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}

	// Load path
	path, err := s.pathRepo.GetByFileID(ctx, file.ID)
	if err == nil {
		file.Path = path
	} else {
		file.Path = []int64{}
	}

	return file, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*File, error) {
	file, err := s.fileRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}
	return file, nil
}

func (s *service) GetRoot(ctx context.Context, ownerID int64) (*File, error) {
	file, err := s.fileRepo.GetByOwnerAndSystemType(ctx, ownerID, SystemTypeRoot)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}
	return file, nil
}

func (s *service) GetHome(ctx context.Context, ownerID int64) (*File, error) {
	file, err := s.fileRepo.GetByOwnerAndSystemType(ctx, ownerID, SystemTypeHome)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrFileNotFound
	}
	return file, nil
}

func (s *service) GetTrash(ctx context.Context, ownerID int64) (*File, error) {
	file, err := s.fileRepo.GetByOwnerAndSystemType(ctx, ownerID, SystemTypeTrash)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, ErrTrashNotFound
	}
	return file, nil
}

func (s *service) GetChildren(ctx context.Context, fileKey string) ([]*File, error) {
	file, err := s.Get(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	if file.Type != FileTypeContainer {
		return nil, ErrNotContainer
	}

	children, err := s.fileRepo.GetChildren(ctx, file.ID)
	if err != nil {
		return nil, err
	}

	return children, nil
}

func (s *service) Create(ctx context.Context, cmd *CreateCommand) (*File, error) {
	// Validate file name
	if cmd.FileName == "" {
		return nil, errors.New("file name is required")
	}

	// Validate file type
	if cmd.Type != FileTypeContainer && cmd.Type != FileTypeFile {
		return nil, errors.New("invalid file type")
	}

	var parentPath []int64

	// Validate parent if specified
	if cmd.ParentKey != nil && *cmd.ParentKey != "" {
		parent, err := s.fileRepo.GetByKey(ctx, *cmd.ParentKey)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrParentNotFound
		}
		if parent.Type != FileTypeContainer {
			return nil, ErrNotContainer
		}
		if parent.OwnerID != cmd.OwnerID {
			return nil, ErrAccessDenied
		}

		// Get parent path
		parentPath, err = s.pathRepo.GetByFileID(ctx, parent.ID)
		if err != nil {
			parentPath = []int64{}
		}
	}

	// Create file entity
	file, err := s.fileRepo.Create(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Create file info
	_, _ = s.infoRepo.Create(ctx, file.ID, 0)

	// Create file path
	newPath := append(parentPath, file.ID)
	_ = s.pathRepo.Create(ctx, file.ID, newPath)

	// Create default role
	_ = s.roleRepo.Create(ctx, cmd.OwnerID, file.ID, []string{"owner", "read", "write"})

	file.Path = newPath
	file.Roles = []string{"owner", "read", "write"}

	return file, nil
}

func (s *service) Update(ctx context.Context, fileKey string, cmd *UpdateCommand) (*File, error) {
	file, err := s.Get(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	// Skip if no updates
	if (cmd.FileName == nil || *cmd.FileName == "") && cmd.ByteSize == nil {
		return file, nil
	}

	updatedFile, err := s.fileRepo.Update(ctx, file.ID, cmd)
	if err != nil {
		return nil, err
	}

	// Update file info if byte size changed
	if cmd.ByteSize != nil {
		_, _ = s.infoRepo.Update(ctx, file.ID, *cmd.ByteSize)
		updatedFile.ByteSize = *cmd.ByteSize
	}

	updatedFile.Path = file.Path
	return updatedFile, nil
}

func (s *service) Delete(ctx context.Context, fileKey string, permanent bool) error {
	file, err := s.Get(ctx, fileKey)
	if err != nil {
		return err
	}

	// Prevent deletion of system files
	if file.IsSystem {
		return ErrCannotDeleteSystem
	}

	if permanent {
		// Permanent delete
		_ = s.infoRepo.Delete(ctx, file.ID)
		_ = s.pathRepo.Delete(ctx, file.ID)
		_ = s.roleRepo.DeleteByFile(ctx, file.ID)
		return s.fileRepo.Delete(ctx, file.ID)
	}

	// Move to trash
	trash, err := s.fileRepo.GetByOwnerAndSystemType(ctx, file.OwnerID, SystemTypeTrash)
	if err != nil {
		return err
	}
	if trash == nil {
		return ErrTrashNotFound
	}

	return s.fileRepo.Move(ctx, file.ID, trash.ID)
}

func (s *service) Move(ctx context.Context, fileKey string, cmd *MoveCommand) (*File, error) {
	file, err := s.Get(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	// Get target container
	target, err := s.fileRepo.GetByKey(ctx, cmd.TargetKey)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, ErrTargetNotFound
	}
	if target.Type != FileTypeContainer {
		return nil, ErrNotContainer
	}
	if target.OwnerID != file.OwnerID {
		return nil, ErrAccessDenied
	}

	// Prevent moving into itself
	if target.ID == file.ID {
		return nil, ErrCannotMoveIntoSelf
	}

	// Move file
	if err := s.fileRepo.Move(ctx, file.ID, target.ID); err != nil {
		return nil, err
	}

	// Update path
	targetPath, err := s.pathRepo.GetByFileID(ctx, target.ID)
	if err != nil {
		targetPath = []int64{}
	}
	newPath := append(targetPath, file.ID)
	_ = s.pathRepo.Update(ctx, file.ID, newPath)

	file.ParentID = &target.ID
	file.Path = newPath
	return file, nil
}

func (s *service) Copy(ctx context.Context, fileKey string, cmd *CopyCommand) (*File, error) {
	file, err := s.Get(ctx, fileKey)
	if err != nil {
		return nil, err
	}

	// Get target container
	target, err := s.fileRepo.GetByKey(ctx, cmd.TargetKey)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, ErrTargetNotFound
	}
	if target.Type != FileTypeContainer {
		return nil, ErrNotContainer
	}
	if target.OwnerID != file.OwnerID {
		return nil, ErrAccessDenied
	}

	// Copy file
	newFile, err := s.fileRepo.Copy(ctx, file.ID, target.ID, file.OwnerID)
	if err != nil {
		return nil, err
	}
	if newFile == nil {
		return nil, errors.New("failed to copy file")
	}

	// Create file info for copy
	_, _ = s.infoRepo.Create(ctx, newFile.ID, file.ByteSize)

	// Create path for copy
	targetPath, err := s.pathRepo.GetByFileID(ctx, target.ID)
	if err != nil {
		targetPath = []int64{}
	}
	newPath := append(targetPath, newFile.ID)
	_ = s.pathRepo.Create(ctx, newFile.ID, newPath)

	// Copy roles
	roles, err := s.roleRepo.GetByUserAndFile(ctx, file.OwnerID, file.ID)
	if err != nil {
		roles = []string{}
	}
	if len(roles) > 0 {
		_ = s.roleRepo.Create(ctx, file.OwnerID, newFile.ID, roles)
	}

	newFile.Path = newPath
	newFile.Roles = roles

	return newFile, nil
}

// EnsureFileKey generates a new UUID for file key if not provided
func EnsureFileKey() string {
	return uuid.New().String()
}
