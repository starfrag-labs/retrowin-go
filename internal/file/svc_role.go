package file

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Role constants
const (
	RoleCreate = "create"
	RoleRead   = "read"
	RoleUpdate = "update"
	RoleDelete = "delete"
	RoleOwner  = "owner"
)

// RoleService defines the interface for file role operations.
type RoleService interface {
	// GetRoles retrieves roles for a user on a file.
	GetRoles(ctx context.Context, userID int64, fileKey string) ([]string, error)

	// CheckRole checks if a user has a specific role on a file.
	CheckRole(ctx context.Context, userID int64, fileKey string, role string) (bool, error)

	// HasReadAccess checks if a user has read access to a file.
	HasReadAccess(ctx context.Context, userID int64, fileKey string) (bool, error)

	// HasWriteAccess checks if a user has write access to a file.
	HasWriteAccess(ctx context.Context, userID int64, fileKey string) (bool, error)

	// HasDeleteAccess checks if a user has delete access to a file.
	HasDeleteAccess(ctx context.Context, userID int64, fileKey string) (bool, error)

	// IsOwner checks if a user is the owner of a file.
	IsOwner(ctx context.Context, userID int64, fileKey string) (bool, error)
}

type roleService struct {
	fileRepo Repository
	roleRepo FileRoleRepository
}

// NewRoleService creates a new file role service.
func NewRoleService(fileRepo Repository, roleRepo FileRoleRepository) RoleService {
	return &roleService{
		fileRepo: fileRepo,
		roleRepo: roleRepo,
	}
}

func (s *roleService) GetRoles(ctx context.Context, userID int64, fileKey string) ([]string, error) {
	file, err := s.fileRepo.GetByKey(ctx, fileKey)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.NotFound("file not found")
	}

	roles, err := s.roleRepo.GetByUserAndFile(ctx, userID, file.ID)
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (s *roleService) CheckRole(ctx context.Context, userID int64, fileKey string, role string) (bool, error) {
	roles, err := s.GetRoles(ctx, userID, fileKey)
	if err != nil {
		return false, err
	}

	for _, r := range roles {
		if r == role || r == RoleOwner {
			return true, nil
		}
	}

	return false, nil
}

func (s *roleService) HasReadAccess(ctx context.Context, userID int64, fileKey string) (bool, error) {
	return s.CheckRole(ctx, userID, fileKey, RoleRead)
}

func (s *roleService) HasWriteAccess(ctx context.Context, userID int64, fileKey string) (bool, error) {
	return s.CheckRole(ctx, userID, fileKey, RoleUpdate)
}

func (s *roleService) HasDeleteAccess(ctx context.Context, userID int64, fileKey string) (bool, error) {
	return s.CheckRole(ctx, userID, fileKey, RoleDelete)
}

func (s *roleService) IsOwner(ctx context.Context, userID int64, fileKey string) (bool, error) {
	file, err := s.fileRepo.GetByKey(ctx, fileKey)
	if err != nil {
		return false, err
	}
	if file == nil {
		return false, errors.NotFound("file not found")
	}

	return file.OwnerID == userID, nil
}
