package file

import (
	"context"
)

// Repository defines the interface for file data access.
type Repository interface {
	// Create creates a new file.
	Create(ctx context.Context, cmd *CreateCommand) (*File, error)

	// GetByID retrieves a file by ID.
	GetByID(ctx context.Context, id int64) (*File, error)

	// GetByKey retrieves a file by file key (UUID).
	GetByKey(ctx context.Context, fileKey string) (*File, error)

	// GetByOwnerAndSystemType retrieves a system file by owner and type.
	GetByOwnerAndSystemType(ctx context.Context, ownerID int64, systemType string) (*File, error)

	// GetChildren retrieves all children of a container.
	GetChildren(ctx context.Context, parentID int64) ([]*File, error)

	// Update updates a file.
	Update(ctx context.Context, id int64, cmd *UpdateCommand) (*File, error)

	// Delete deletes a file by ID.
	Delete(ctx context.Context, id int64) error

	// ExistsByKey checks if a file exists by key.
	ExistsByKey(ctx context.Context, fileKey string) (bool, error)

	// GetByOwnerAndParent retrieves files by owner and parent.
	GetByOwnerAndParent(ctx context.Context, ownerID int64, parentID *int64) ([]*File, error)

	// UpdateByteSize updates the byte size of a file.
	UpdateByteSize(ctx context.Context, id int64, byteSize int64) error

	// Move moves a file to a new parent.
	Move(ctx context.Context, fileID int64, newParentID int64) error

	// Copy copies a file to a new parent and returns the new file.
	Copy(ctx context.Context, fileID int64, newParentID int64, ownerID int64) (*File, error)
}

// FileInfoRepository defines the interface for file info data access.
type FileInfoRepository interface {
	// Create creates file info for a file.
	Create(ctx context.Context, fileID int64, byteSize int64) (*FileInfo, error)

	// GetByFileID retrieves file info by file ID.
	GetByFileID(ctx context.Context, fileID int64) (*FileInfo, error)

	// Update updates file info.
	Update(ctx context.Context, fileID int64, byteSize int64) (*FileInfo, error)

	// Delete deletes file info by file ID.
	Delete(ctx context.Context, fileID int64) error
}

// FilePathRepository defines the interface for file path data access.
type FilePathRepository interface {
	// Create creates a file path for a file.
	Create(ctx context.Context, fileID int64, path []int64) error

	// GetByFileID retrieves the file path by file ID.
	GetByFileID(ctx context.Context, fileID int64) ([]int64, error)

	// Update updates file path.
	Update(ctx context.Context, fileID int64, path []int64) error

	// Delete deletes file path by file ID.
	Delete(ctx context.Context, fileID int64) error
}

// FileRoleRepository defines the interface for file role data access.
type FileRoleRepository interface {
	// Create creates a file role.
	Create(ctx context.Context, userID int64, fileID int64, roles []string) error

	// GetByUserAndFile retrieves roles for a user on a file.
	GetByUserAndFile(ctx context.Context, userID int64, fileID int64) ([]string, error)

	// Update updates file roles.
	Update(ctx context.Context, userID int64, fileID int64, roles []string) error

	// Delete deletes file roles.
	Delete(ctx context.Context, userID int64, fileID int64) error

	// DeleteByFile deletes all roles for a file.
	DeleteByFile(ctx context.Context, fileID int64) error
}
