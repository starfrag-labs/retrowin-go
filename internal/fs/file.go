package fs

import (
	"github.com/starfrag-lab/retrowin-go/internal/filedata"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
	"github.com/starfrag-lab/retrowin-go/internal/symlink"
)

// File represents a file with path context (aggregate of inode + entry + data).
type File struct {
	*inode.Inode
	Name          string             `json:"name"`
	Path          string             `json:"path"`
	ParentID      *int64             `json:"parentId,omitempty"`
	FileData      *filedata.FileData `json:"fileData,omitempty"`
	SymlinkTarget *symlink.Symlink   `json:"symlink,omitempty"`
}
