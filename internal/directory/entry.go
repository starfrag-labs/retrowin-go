package directory

// Entry represents a name → inode mapping in a directory.
type Entry struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parentId"`
	Name     string `json:"name"`
	ChildID  int64  `json:"childId"`
}
