package fs

import (
	"context"
	"slices"

	"github.com/starfrag-lab/retrowin-go/internal/core/fs/etc"
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

// checkPerm verifies if uid has the requested access to the inode.
func (s *service) checkPerm(in *inode.Inode, uid int, access AccessType) error {
	if uid == 0 {
		return nil // uid not set, skip permission check (e.g. internal calls)
	}

	gids := s.resolveGIDs(context.Background(), in.SystemID(), uid)
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

// resolveGIDs returns all gids the given uid belongs to in the system.
// TODO: read /etc/group inode from filesystem.
func (s *service) resolveGIDs(_ context.Context, _ string, _ int) []int {
	return nil
}

// resolveGIDsFromContent parses /etc/group content for the given uid.
func resolveGIDsFromContent(data []byte, uid int) []int {
	return etc.ResolveGIDsByUID(data, uid)
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
