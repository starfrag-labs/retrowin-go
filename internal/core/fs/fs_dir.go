package fs

import (
	"context"
	"encoding/json"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
)

// Link adds a directory entry to a directory inode.
func (s *service) Link(ctx context.Context, dirID string, entry content.DirEntry) error {
	dir, err := s.inodeSvc.GetByID(ctx, dirID)
	if err != nil {
		return err
	}
	if !dir.IsDir() {
		return errors.BadRequest("not a directory")
	}

	var c content.DirContent
	if dir.Content() != nil {
		if err := json.Unmarshal(dir.Content(), &c); err != nil {
			return errors.WrapInternal(err, "failed to parse directory content")
		}
	}

	// Check for duplicate name
	for _, e := range c.Entries {
		if e.Name == entry.Name {
			return errors.Conflict("entry already exists: " + entry.Name)
		}
	}

	c.Entries = append(c.Entries, entry)
	raw, err := json.Marshal(c)
	if err != nil {
		return errors.WrapInternal(err, "failed to marshal directory content")
	}

	return s.inodeSvc.Update(ctx, &inode.UpdateCommand{
		ID:      dirID,
		Content: &raw,
	})
}

// Unlink removes a directory entry from a directory inode.
func (s *service) Unlink(ctx context.Context, dirID string, name string) error {
	dir, err := s.inodeSvc.GetByID(ctx, dirID)
	if err != nil {
		return err
	}
	if !dir.IsDir() {
		return errors.BadRequest("not a directory")
	}

	var c content.DirContent
	if dir.Content() == nil {
		return nil
	}
	if err := json.Unmarshal(dir.Content(), &c); err != nil {
		return errors.WrapInternal(err, "failed to parse directory content")
	}

	filtered := make([]content.DirEntry, 0, len(c.Entries))
	found := false
	for _, e := range c.Entries {
		if e.Name == name {
			found = true
			continue
		}
		filtered = append(filtered, e)
	}
	if !found {
		return nil
	}

	c.Entries = filtered
	raw, err := json.Marshal(c)
	if err != nil {
		return errors.WrapInternal(err, "failed to marshal directory content")
	}

	return s.inodeSvc.Update(ctx, &inode.UpdateCommand{
		ID:      dirID,
		Content: &raw,
	})
}
