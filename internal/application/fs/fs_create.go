package fs

import (
	"context"
	"encoding/json"

	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

func (s *service) CreateFile(ctx context.Context, cmd *CreateFileCommand) (*inode.Inode, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}

	uid, gids, err := s.userSvc.ResolveUIDAndGIDs(ctx, cmd.SystemID)
	if err != nil {
		return nil, err
	}

	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeRegular | inode.PermOwnerRW | inode.PermGroupRX | inode.PermOtherR
	}

	gid := cmd.GID
	if gid == 0 && len(gids) > 0 {
		gid = gids[0]
	}

	return s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
		UID:      uid,
		GID:      gid,
		Size:     cmd.Size,
		Flags:    cmd.Flags,
		Content:  cmd.Content,
	})
}

func (s *service) CreateDirectory(ctx context.Context, cmd *CreateDirectoryCommand) (*inode.Inode, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}

	uid, gids, err := s.userSvc.ResolveUIDAndGIDs(ctx, cmd.SystemID)
	if err != nil {
		return nil, err
	}

	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeDirectory | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherR
	}

	gid := cmd.GID
	if gid == 0 && len(gids) > 0 {
		gid = gids[0]
	}

	dirContent := content.DirContent{Entries: []content.DirEntry{}}
	raw, err := json.Marshal(dirContent)
	if err != nil {
		return nil, err
	}

	return s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
		UID:      uid,
		GID:      gid,
		Flags:    cmd.Flags,
		Content:  raw,
	})
}

func (s *service) CreateSymlink(ctx context.Context, cmd *CreateSymlinkCommand) (*inode.Inode, error) {
	if cmd.SystemID == "" {
		return nil, errors.BadRequest("system_id is required")
	}
	if cmd.Target == "" {
		return nil, errors.BadRequest("target is required")
	}

	uid, gids, err := s.userSvc.ResolveUIDAndGIDs(ctx, cmd.SystemID)
	if err != nil {
		return nil, err
	}

	mode := cmd.Mode
	if mode == 0 {
		mode = inode.ModeSymlink | inode.PermOwnerRWX | inode.PermGroupRX | inode.PermOtherR
	}

	gid := cmd.GID
	if gid == 0 && len(gids) > 0 {
		gid = gids[0]
	}

	symContent := content.SymlinkContent{Target: cmd.Target}
	raw, err := json.Marshal(symContent)
	if err != nil {
		return nil, err
	}

	return s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: cmd.SystemID,
		Mode:     mode,
		UID:      uid,
		GID:      gid,
		Flags:    cmd.Flags,
		Content:  raw,
	})
}
