package content

// DirContent is stored in inode content for ModeDirectory inodes.
type DirContent struct {
	Entries []DirEntry `json:"entries"`
}

// DirEntry represents a filename to inode mapping within a directory.
type DirEntry struct {
	Name     string `json:"name"`
	InodeID  string `json:"inode_id"`
	FileType uint8  `json:"file_type"`
}
