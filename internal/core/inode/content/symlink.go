package content

// SymlinkContent is stored in inode content for ModeSymlink inodes.
type SymlinkContent struct {
	Target string `json:"target"`
}
