package fs

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

func (s *service) Get(ctx context.Context, id string) (*inode.Inode, error) {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkPermFromContext(ctx, in, AccessRead); err != nil {
		return nil, err
	}

	return in, nil
}

func (s *service) List(ctx context.Context, filter *ListFilter) ([]*inode.Inode, error) {
	f := inode.Filter{}
	if filter.SystemID != nil {
		f = inode.BySystemID(*filter.SystemID)
	}
	if filter.UID != nil {
		f.UID = filter.UID
	}
	return s.inodeSvc.Find(ctx, f)
}

func (s *service) ReadDir(ctx context.Context, id string) ([]content.DirEntry, error) {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !in.IsDir() {
		return nil, errors.BadRequest("not a directory")
	}

	if err := s.checkPermFromContext(ctx, in, AccessRead); err != nil {
		return nil, err
	}

	if in.Content() == nil {
		return nil, nil
	}

	var c content.DirContent
	if err := json.Unmarshal(in.Content(), &c); err != nil {
		return nil, errors.WrapInternal(err, "failed to parse directory content")
	}

	if len(c.Entries) == 0 {
		return c.Entries, nil
	}

	// Validate InodeID existence and filter out dangling entries
	validEntries := make([]content.DirEntry, 0, len(c.Entries))
	for _, entry := range c.Entries {
		if _, err := s.inodeSvc.GetByID(ctx, entry.InodeID); err != nil {
			slog.Warn("dangling directory entry found, filtering out",
				"entry_name", entry.Name,
				"inode_id", entry.InodeID,
				"parent_dir_id", id,
			)
			continue
		}
		validEntries = append(validEntries, entry)
	}

	// Auto-clean parent if dangling entries were removed
	if len(validEntries) != len(c.Entries) {
		c.Entries = validEntries
		raw, err := json.Marshal(c)
		if err != nil {
			slog.Error("failed to marshal cleaned directory content", "error", err)
		} else if err := s.inodeSvc.Update(ctx, &inode.UpdateCommand{
			ID:      id,
			Content: &raw,
		}); err != nil {
			slog.Error("failed to clean up dangling directory entries", "error", err)
		}
	}

	return validEntries, nil
}
