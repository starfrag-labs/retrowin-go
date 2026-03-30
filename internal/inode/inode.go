package inode

import (
	"time"
)

// FileType represents the type of an inode.
type FileType string

const (
	FileTypeRegular   FileType = "regular"
	FileTypeDirectory FileType = "directory"
	FileTypeSymlink   FileType = "symlink"
	FileTypeBlock     FileType = "block"
	FileTypeChar      FileType = "char"
	FileTypeSocket    FileType = "socket"
	FileTypeFifo      FileType = "fifo"
)

// System file types
const (
	SystemTypeRoot  = "root"
	SystemTypeHome  = "home"
	SystemTypeTrash = "trash"
)

// Inode represents a file system inode (metadata only, no filename).
type Inode struct {
	id         int64
	systemID   *int64
	fileType   FileType
	byteSize   int64
	ownerUID   string
	ownerGID   string
	permOwner  string
	permGroup  string
	permOthers string
	linkCount  int16
	accessedAt *time.Time
	isSystem   bool
	systemType *string
	createdAt  time.Time
	updatedAt  time.Time
}

// NewInode creates a new Inode.
func NewInode(
	id int64,
	systemID *int64,
	fileType FileType,
	byteSize int64,
	ownerUID string,
	ownerGID string,
	permOwner string,
	permGroup string,
	permOthers string,
	linkCount int16,
	accessedAt *time.Time,
	isSystem bool,
	systemType *string,
	createdAt time.Time,
	updatedAt time.Time,
) *Inode {
	return &Inode{
		id:         id,
		systemID:   systemID,
		fileType:   fileType,
		byteSize:   byteSize,
		ownerUID:   ownerUID,
		ownerGID:   ownerGID,
		permOwner:  permOwner,
		permGroup:  permGroup,
		permOthers: permOthers,
		linkCount:  linkCount,
		accessedAt: accessedAt,
		isSystem:   isSystem,
		systemType: systemType,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

// Getters
func (i *Inode) ID() int64              { return i.id }
func (i *Inode) SystemID() *int64       { return i.systemID }
func (i *Inode) FileType() FileType     { return i.fileType }
func (i *Inode) ByteSize() int64        { return i.byteSize }
func (i *Inode) OwnerUID() string       { return i.ownerUID }
func (i *Inode) OwnerGID() string       { return i.ownerGID }
func (i *Inode) PermOwner() string      { return i.permOwner }
func (i *Inode) PermGroup() string      { return i.permGroup }
func (i *Inode) PermOthers() string     { return i.permOthers }
func (i *Inode) LinkCount() int16       { return i.linkCount }
func (i *Inode) AccessedAt() *time.Time { return i.accessedAt }
func (i *Inode) IsSystem() bool         { return i.isSystem }
func (i *Inode) SystemType() *string    { return i.systemType }
func (i *Inode) CreatedAt() time.Time   { return i.createdAt }
func (i *Inode) UpdatedAt() time.Time   { return i.updatedAt }
