package fs

import (
	"context"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
)

func (s *service) Copy(ctx context.Context, id string, systemID string) (*inode.Inode, error) {
	original, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Object inodes reference external storage - cannot simply copy
	if original.IsObject() {
		return nil, errors.BadRequest("cannot copy object inode: use upload service instead")
	}

	return s.inodeSvc.Create(ctx, &inode.CreateCommand{
		SystemID: systemID,
		Mode:     original.Mode(),
		UID:      original.UID(),
		GID:      original.GID(),
		Flags:    original.Flags(),
		Content:  original.Content(),
	})
}
