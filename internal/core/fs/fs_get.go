package fs

import (
	"context"
	"encoding/json"

	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

func (s *service) Get(ctx context.Context, uid int, id string) (*inode.Inode, error) {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkPerm(in, uid, AccessRead); err != nil {
		return nil, err
	}

	return in, nil
}

func (s *service) List(ctx context.Context, uid int, filter *ListFilter) ([]*inode.Inode, error) {
	f := inode.Filter{}
	if filter.SystemID != nil {
		f = inode.BySystemID(*filter.SystemID)
	}
	if filter.UID != nil {
		f.UID = filter.UID
	}
	return s.inodeSvc.Find(ctx, f)
}

func (s *service) ReadDir(ctx context.Context, uid int, id string) ([]content.DirEntry, error) {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !in.IsDir() {
		return nil, errors.BadRequest("not a directory")
	}

	if err := s.checkPerm(in, uid, AccessRead); err != nil {
		return nil, err
	}

	if in.Content() == nil {
		return nil, nil
	}

	var c content.DirContent
	if err := json.Unmarshal(in.Content(), &c); err != nil {
		return nil, errors.WrapInternal(err, "failed to parse directory content")
	}
	return c.Entries, nil
}
