package inode

import (
	"time"
)

// Mode constants following Linux inode conventions.
// The mode field encodes both file type (upper bits) and permissions (lower bits).
const (
	// File types (mode & 0xF000)
	ModeTypeMask  = 0xF000
	ModeRegular   = 0x8000 // regular file
	ModeDirectory = 0x4000 // directory
	ModeSymlink   = 0xA000 // symbolic link
	ModeBlock     = 0x6000 // block device
	ModeChar      = 0x2000 // character device
	ModeFifo      = 0x1000 // FIFO
	ModeSocket    = 0xC000 // socket

	// Permission bits (mode & 0x0FFF)
	PermOwnerRead  = 0x0100
	PermOwnerWrite = 0x0080
	PermOwnerExec  = 0x0040
	PermGroupRead  = 0x0020
	PermGroupWrite = 0x0010
	PermGroupExec  = 0x0008
	PermOtherRead  = 0x0004
	PermOtherWrite = 0x0002
	PermOtherExec  = 0x0001

	// Common permission combinations
	PermOwnerRWX = PermOwnerRead | PermOwnerWrite | PermOwnerExec
	PermGroupRX  = PermGroupRead | PermGroupExec
	PermOtherR   = PermOtherRead
	PermOwnerRW  = PermOwnerRead | PermOwnerWrite
)

// Inode represents a file system inode (metadata only, no filename).
// Follows Linux inode structure: mode, uid, gid, size, timestamps.
type Inode struct {
	id        int64
	systemID  string
	mode      int
	uid       int64
	gid       int64
	size      int64
	linkCount int
	flags     int
	atime     time.Time
	mtime     time.Time
	ctime     time.Time
	content   []byte
	createdAt time.Time
	updatedAt time.Time
}

// NewInode creates a new Inode.
func NewInode(
	id int64,
	systemID string,
	mode int,
	uid int64,
	gid int64,
	size int64,
	linkCount int,
	flags int,
	atime time.Time,
	mtime time.Time,
	ctime time.Time,
	content []byte,
	createdAt time.Time,
	updatedAt time.Time,
) *Inode {
	return &Inode{
		id:        id,
		systemID:  systemID,
		mode:      mode,
		uid:       uid,
		gid:       gid,
		size:      size,
		linkCount: linkCount,
		flags:     flags,
		atime:     atime,
		mtime:     mtime,
		ctime:     ctime,
		content:   content,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

// Getters
func (i *Inode) ID() int64            { return i.id }
func (i *Inode) SystemID() string     { return i.systemID }
func (i *Inode) Mode() int            { return i.mode }
func (i *Inode) UID() int64           { return i.uid }
func (i *Inode) GID() int64           { return i.gid }
func (i *Inode) Size() int64          { return i.size }
func (i *Inode) LinkCount() int       { return i.linkCount }
func (i *Inode) Flags() int           { return i.flags }
func (i *Inode) Atime() time.Time     { return i.atime }
func (i *Inode) Mtime() time.Time     { return i.mtime }
func (i *Inode) Ctime() time.Time     { return i.ctime }
func (i *Inode) Content() []byte      { return i.content }
func (i *Inode) CreatedAt() time.Time { return i.createdAt }
func (i *Inode) UpdatedAt() time.Time { return i.updatedAt }

// FileType returns the file type portion of the mode.
func (i *Inode) FileType() int {
	return i.mode & ModeTypeMask
}

// Permissions returns the permission portion of the mode.
func (i *Inode) Permissions() int {
	return i.mode & 0x0FFF
}

// IsDir returns true if the inode represents a directory.
func (i *Inode) IsDir() bool {
	return i.FileType() == ModeDirectory
}

// IsRegular returns true if the inode represents a regular file.
func (i *Inode) IsRegular() bool {
	return i.FileType() == ModeRegular
}

// IsSymlink returns true if the inode represents a symbolic link.
func (i *Inode) IsSymlink() bool {
	return i.FileType() == ModeSymlink
}
