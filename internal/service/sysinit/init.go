package sysinit

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/core/user"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/system"
)

// InitService handles system initialization.
type InitService interface {
	// InitSystem creates a new system with initial users and directories.
	// Returns the created system, root user, and initial user.
	InitSystem(ctx context.Context, cmd *InitSystemCommand) (*InitResult, error)
}

// InitSystemCommand for initializing a new system.
type InitSystemCommand struct {
	Name         string
	Description  *string
	RootUserID   string   // External user ID for root
	InitialUsers []InitialUser // Optional initial users to create
}

// InitialUser represents an initial user to create.
type InitialUser struct {
	UserID   string
	Username string
}

// InitResult contains the result of system initialization.
type InitResult struct {
	System       *system.System
	RootUser     *user.SystemUser
	InitialUsers []*user.SystemUser
	RootDir      *inode.Inode
	HomeDir      *inode.Inode
}

type service struct {
	systemSvc system.SystemService
	userSvc   user.UserService
	fsSvc     fs.FsService
}

// NewService creates a new init service.
func NewService(systemSvc system.SystemService, userSvc user.UserService, fsSvc fs.FsService) InitService {
	return &service{
		systemSvc: systemSvc,
		userSvc:   userSvc,
		fsSvc:     fsSvc,
	}
}

// InitSystem creates a new system with initial users and directories.
func (s *service) InitSystem(ctx context.Context, cmd *InitSystemCommand) (*InitResult, error) {
	if cmd.Name == "" {
		return nil, errors.BadRequest("name is required")
	}

	// 1. Create the system
	sys, err := s.systemSvc.Create(ctx, &system.CreateCommand{
		Name:        cmd.Name,
		Description: cmd.Description,
		Status:      system.StatusActive,
	})
	if err != nil {
		return nil, err
	}

	// 2. Create root user (UID=0, GID=0)
	rootUser, err := s.createRootUser(ctx, sys.ID(), cmd.RootUserID)
	if err != nil {
		return nil, err
	}

	// 3. Create initial users
	initialUsers := make([]*user.SystemUser, 0, len(cmd.InitialUsers))
	for _, u := range cmd.InitialUsers {
		sysUser, err := s.userSvc.Create(ctx, &user.CreateCommand{
			UserID:   u.UserID,
			SystemID: sys.ID(),
			Username: u.Username,
		})
		if err != nil {
			return nil, err
		}
		initialUsers = append(initialUsers, sysUser)
	}

	// 4. Create root directory (/) - owned by root
	rootDir, err := s.createRootDirectory(ctx, sys.ID())
	if err != nil {
		return nil, err
	}

	// 5. Create /home directory - owned by root
	homeDir, err := s.createHomeDirectory(ctx, sys.ID(), rootDir.ID())
	if err != nil {
		return nil, err
	}

	return &InitResult{
		System:       sys,
		RootUser:     rootUser,
		InitialUsers: initialUsers,
		RootDir:      rootDir,
		HomeDir:      homeDir,
	}, nil
}

// createRootUser creates the root user with UID=0.
func (s *service) createRootUser(ctx context.Context, systemID, rootUserID string) (*user.SystemUser, error) {
	// Create root user with explicit UID=0
	return s.userSvc.Create(ctx, &user.CreateCommand{
		UserID:   rootUserID,
		SystemID: systemID,
		Username: "root",
		UID:      0, // Root UID
	})
}

// createRootDirectory creates the root directory (/).
func (s *service) createRootDirectory(ctx context.Context, systemID string) (*inode.Inode, error) {
	rootDir, err := s.fsSvc.CreateDirectory(ctx, &fs.CreateDirectoryCommand{
		SystemID: systemID,
		GID:      0, // root group
		Mode:     inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherRX,
		Flags:    inode.FlagRoot, // Mark as root directory
	})
	if err != nil {
		return nil, err
	}
	return rootDir, nil
}

// createHomeDirectory creates /home directory under root.
func (s *service) createHomeDirectory(ctx context.Context, systemID string, rootDirID string) (*inode.Inode, error) {
	// Create /home directory
	homeDir, err := s.fsSvc.CreateDirectory(ctx, &fs.CreateDirectoryCommand{
		SystemID: systemID,
		GID:      0, // root group
		Mode:     inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherRX,
	})
	if err != nil {
		return nil, err
	}

	// Link home to root directory
	if err := s.fsSvc.Link(ctx, rootDirID, content.DirEntry{
		Name:     "home",
		InodeID:  homeDir.ID(),
		FileType: uint8(inode.ModeDirectory >> 8),
	}); err != nil {
		return nil, err
	}

	return homeDir, nil
}
