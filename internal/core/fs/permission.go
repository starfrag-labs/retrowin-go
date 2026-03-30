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

// checkPermFromContext extracts user info from context and checks permissions.
func (s *service) checkPermFromContext(ctx context.Context, in *inode.Inode, access AccessType) error {
	uid, err := s.userSvc.ResolveUID(ctx, in.SystemID())
	if err != nil {
		return err
	}
	if uid == 0 {
		return nil // No user context, skip permission check (internal/system calls)
	}

	gids := s.resolveGIDs(ctx, in.SystemID(), uid)
	return s.checkPermWithGIDs(in, uid, gids, access)
}

// checkPerm verifies if uid has the requested access to the inode.
// Used for internal operations where uid is already known.
func (s *service) checkPerm(in *inode.Inode, uid int, access AccessType) error {
	if uid == 0 {
		return nil // uid not set, skip permission check (e.g. internal calls)
	}

	gids := s.resolveGIDs(context.Background(), in.SystemID(), uid)
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

// resolveGIDs returns all gids the given uid belongs to in the system.
// Reads /etc/group from the filesystem.
func (s *service) resolveGIDs(ctx context.Context, systemID string, uid int) []int {
	groupContent, err := s.getEtcGroupContent(ctx, systemID)
	if err != nil {
		return nil
	}
	return resolveGIDsFromContent(groupContent, uid)
}

// getEtcGroupContent retrieves /etc/group file content from the filesystem.
func (s *service) getEtcGroupContent(ctx context.Context, systemID string) ([]byte, error) {
	// TODO: Implement /etc/group lookup from filesystem
	// For now, return nil to skip group-based permissions
	// This should look up a well-known inode ID or path for /etc/group
	return nil, nil
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
