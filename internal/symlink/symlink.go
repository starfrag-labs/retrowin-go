package symlink

// Symlink represents a symbolic link target.
type Symlink struct {
	inodeID    int64
	targetPath string
}

// NewSymlink creates a new Symlink.
func NewSymlink(inodeID int64, targetPath string) *Symlink {
	return &Symlink{
		inodeID:    inodeID,
		targetPath: targetPath,
	}
}

// Getters
func (s *Symlink) InodeID() int64     { return s.inodeID }
func (s *Symlink) TargetPath() string { return s.targetPath }
