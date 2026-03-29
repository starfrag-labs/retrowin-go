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

// StorageType represents the storage backend type.
type StorageType string

const (
	StorageTypeS3    StorageType = "s3"
	StorageTypeLocal StorageType = "local"
)

// System file types
const (
	SystemTypeRoot  = "root"
	SystemTypeHome  = "home"
	SystemTypeTrash = "trash"
)

// Inode represents a file system inode (metadata only, no filename).
type Inode struct {
	ID          int64      `json:"id"`
	SystemID    *int64     `json:"systemId,omitempty"`
	FileType    FileType   `json:"fileType"`
	ByteSize    int64      `json:"byteSize"`
	OwnerUID    string     `json:"ownerUid"`
	OwnerGID    string     `json:"ownerGid"`
	PermOwner   string     `json:"permOwner"`
	PermGroup   string     `json:"permGroup"`
	PermOthers  string     `json:"permOthers"`
	LinkCount   int16      `json:"linkCount"`
	AccessedAt  *time.Time `json:"accessedAt,omitempty"`
	IsSystem    bool       `json:"isSystem"`
	SystemType  *string    `json:"systemType,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// FileData represents the storage info for a regular file.
// This is an internal detail, not exposed as a service.
type FileData struct {
	InodeID     int64       `json:"inodeId"`
	StorageType StorageType `json:"storageType"`
	Location    string      `json:"location"`
	Checksum    *string     `json:"checksum,omitempty"`
}
