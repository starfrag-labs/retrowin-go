package symlink

// Symlink represents a symbolic link target.
type Symlink struct {
	InodeID    int64  `json:"inodeId"`
	TargetPath string `json:"targetPath"`
}
