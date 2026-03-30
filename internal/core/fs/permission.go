package fs

import (
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

// CheckPermission verifies if a user has the requested access to an inode.
// Returns nil if access is granted, or an error if denied.
func CheckPermission(in *inode.Inode, uid, gid int, access AccessType) error {
	mode := in.Permissions()

	var perm int
	switch {
	case in.UID() == uid:
		perm = ownerPerm(mode, access)
	case in.GID() == gid:
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
