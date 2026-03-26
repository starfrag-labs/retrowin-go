package file

import (
	"time"
)

// FileType represents the type of a file.
type FileType string

const (
	FileTypeContainer FileType = "container"
	FileTypeFile      FileType = "file"
)

// File represents a file or container in the system.
type File struct {
	ID         int64     `json:"id"`
	FileKey    string    `json:"fileKey"`
	Type       FileType  `json:"type"`
	FileName   string    `json:"fileName"`
	OwnerID    int64     `json:"ownerId"`
	ParentID   *int64    `json:"parentId"`
	ByteSize   int64     `json:"byteSize"`
	IsSystem   bool      `json:"isSystem"`
	SystemType *string   `json:"systemType"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Path       []int64   `json:"path"`
	Roles      []string  `json:"roles"`
}

// System file types
const (
	SystemTypeRoot  = "root"
	SystemTypeHome  = "home"
	SystemTypeTrash = "trash"
)

// FileInfo represents additional file metadata.
type FileInfo struct {
	FileID     int64     `json:"fileId"`
	CreateDate time.Time `json:"createDate"`
	UpdateDate time.Time `json:"updateDate"`
	ByteSize   int64     `json:"byteSize"`
}

// CreateCommand represents the command to create a new file.
type CreateCommand struct {
	Type      FileType `json:"type"`
	FileName  string   `json:"fileName"`
	ParentKey *string  `json:"parentKey"`
	OwnerID   int64    `json:"ownerId"`
}

// UpdateCommand represents the command to update a file.
type UpdateCommand struct {
	FileName *string `json:"fileName"`
	ByteSize *int64  `json:"byteSize"`
}

// MoveCommand represents the command to move a file.
type MoveCommand struct {
	TargetKey string `json:"targetKey"`
}

// CopyCommand represents the command to copy a file.
type CopyCommand struct {
	TargetKey string `json:"targetKey"`
}
