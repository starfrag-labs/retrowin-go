package fs

import (
	"context"
	"slices"

	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// AccessType represents the type of access being requested.
type AccessType int

const (
	AccessRead AccessType = iota
	AccessWrite
	AccessExecute
)

// checkPermFromContext extracts user info from context and checks permissions.
func (s *service) checkPermFromContext(ctx context.Context, in *inode.Inode, access AccessType) error {
	uid, gids, err := s.userSvc.ResolveUIDAndGIDs(ctx, in.SystemID())
	if err != nil {
		return err
	}
	if uid == 0 {
		return nil // No user context, skip permission check (internal/system calls)
	}

	return s.checkPermWithGIDs(in, uid, gids, access)
}

// checkPermWithGIDs performs the actual permission check.
func (s *service) checkPermWithGIDs(in *inode.Inode, uid int, gids []int, access AccessType) error {
	mode := in.Permissions()

	var perm int
	switch {
	case in.UID() == uid:
		perm = ownerPerm(mode, access)
	case slices.Contains(gids, in.GID()):
		perm = groupPerm(mode, access)
	default:
		perm = otherPerm(mode, access)
	}

	if perm == 0 {
		return errors.Forbidden("permission denied")
	}
	return nil
}

func ownerPerm(mode int, access AccessType) int {
	switch access {
	case AccessRead:
		return mode & inode.PermOwnerRead
	case AccessWrite:
		return mode & inode.PermOwnerWrite
	case AccessExecute:
		return mode & inode.PermOwnerExec
	}
	return 0
}

func groupPerm(mode int, access AccessType) int {
	switch access {
	case AccessRead:
		return mode & inode.PermGroupRead
	case AccessWrite:
		return mode & inode.PermGroupWrite
	case AccessExecute:
		return mode & inode.PermGroupExec
	}
	return 0
}

func otherPerm(mode int, access AccessType) int {
	switch access {
	case AccessRead:
		return mode & inode.PermOtherRead
	case AccessWrite:
		return mode & inode.PermOtherWrite
	case AccessExecute:
		return mode & inode.PermOtherExec
	}
	return 0
}
