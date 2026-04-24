package fs

import (
	"context"
	"path"

	"github.com/starfrag-lab/retrowin-go/internal/core/dentry"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// Mkdir creates a directory at the given path.
func (s *service) Mkdir(ctx context.Context, systemID string, filePath string, mode int) (*inode.Inode, error) {
	dirPath := path.Dir(filePath)
	dirName := path.Base(filePath)

	if dirPath == "/" && dirName == "/" {
		return nil, errors.BadRequest("cannot create root directory")
	}

	parentDir, err := s.ResolvePath(ctx, systemID, dirPath)
	if err != nil {
		return nil, err
	}

	if mode == 0 {
		mode = inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherRX
	}

	dirInode, err := s.CreateDirectory(ctx, &CreateDirectoryCommand{
		SystemID: systemID,
		Mode:     mode,
	})
	if err != nil {
		return nil, err
	}

	entry := dentry.DirEntry{
		Name:     dirName,
		InodeID:  dirInode.ID(),
		FileType: uint8(inode.ModeDirectory >> 12),
	}
	if err := s.dentrySvc.Link(ctx, parentDir.ID(), entry); err != nil {
		return nil, err
	}

	return dirInode, nil
}

// Ln creates a symbolic link at linkPath pointing to target.
func (s *service) Ln(ctx context.Context, systemID string, linkPath string, target string) (*inode.Inode, error) {
	if target == "" {
		return nil, errors.BadRequest("target path is required")
	}
	if len(target) > 4096 {
		return nil, errors.BadRequest("target path too long")
	}

	linkDir := path.Dir(linkPath)
	linkName := path.Base(linkPath)

	parentDir, err := s.ResolvePath(ctx, systemID, linkDir)
	if err != nil {
		return nil, err
	}

	symlinkInode, err := s.CreateSymlink(ctx, &CreateSymlinkCommand{
		SystemID: systemID,
		Target:   target,
		Mode:     inode.ModeSymlink | 0777,
	})
	if err != nil {
		return nil, err
	}

	entry := dentry.DirEntry{
		Name:     linkName,
		InodeID:  symlinkInode.ID(),
		FileType: uint8(inode.ModeSymlink >> 12),
	}
	if err := s.dentrySvc.Link(ctx, parentDir.ID(), entry); err != nil {
		return nil, err
	}

	return symlinkInode, nil
}

// UnlinkPath removes a path (file, symlink, or empty directory).
func (s *service) UnlinkPath(ctx context.Context, systemID string, filePath string) error {
	dirPath := path.Dir(filePath)
	entryName := path.Base(filePath)

	parentDir, err := s.ResolvePath(ctx, systemID, dirPath)
	if err != nil {
		return err
	}

	entries, err := s.dentrySvc.ReadDir(ctx, parentDir.ID())
	if err != nil {
		return err
	}

	var targetEntry *dentry.DirEntry
	for i := range entries {
		if entries[i].Name == entryName {
			targetEntry = &entries[i]
			break
		}
	}
	if targetEntry == nil {
		return errors.NotFound("path not found: " + filePath)
	}

	if err := s.dentrySvc.Unlink(ctx, parentDir.ID(), entryName); err != nil {
		return err
	}

	if err := s.Delete(ctx, targetEntry.InodeID); err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}

// ChmodPath changes permissions of a path.
func (s *service) ChmodPath(ctx context.Context, systemID string, filePath string, mode int) (*inode.Inode, error) {
	if mode < 0 || mode > 0o777 {
		return nil, errors.BadRequest("mode must be between 0 and 0o777")
	}

	in, err := s.ResolvePath(ctx, systemID, filePath)
	if err != nil {
		return nil, err
	}

	newMode := (in.Mode() & inode.ModeTypeMask) | mode
	if err := s.UpdateMode(ctx, &UpdateModeCommand{
		ID:   in.ID(),
		Mode: newMode,
	}); err != nil {
		return nil, err
	}

	return s.Get(ctx, in.ID())
}
