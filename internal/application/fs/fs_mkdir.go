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
