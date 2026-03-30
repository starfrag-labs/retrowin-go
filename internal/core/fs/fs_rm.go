package fs

import (
	"context"
	"encoding/json"

	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

func (s *service) Delete(ctx context.Context, id string) error {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.checkPermFromContext(ctx, in, AccessWrite); err != nil {
		return err
	}

	switch {
	case in.IsObject():
		if err := s.deleteObjectRef(ctx, in.Content()); err != nil {
			return err
		}
	case in.IsDir():
		if err := s.ensureDirEmpty(in.Content()); err != nil {
			return err
		}
	}

	return s.inodeSvc.Delete(ctx, id)
}

func (s *service) deleteObjectRef(ctx context.Context, raw []byte) error {
	if raw == nil {
		return nil
	}
	var c content.ObjectContent
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil
	}
	if c.ObjectID == "" {
		return nil
	}
	if err := s.objectSvc.Delete(ctx, c.ObjectID); err != nil {
		return errors.WrapInternal(err, "failed to delete object")
	}
	return nil
}

func (s *service) ensureDirEmpty(raw []byte) error {
	if raw == nil {
		return nil
	}
	var c content.DirContent
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil
	}
	if len(c.Entries) > 0 {
		return errors.BadRequest("directory not empty")
	}
	return nil
}
