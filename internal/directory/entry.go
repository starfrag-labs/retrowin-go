package directory

// Entry represents a name → inode mapping in a directory.
type Entry struct {
	id       int64
	parentID int64
	name     string
	childID  int64
}

// NewEntry creates a new Entry.
func NewEntry(id int64, parentID int64, name string, childID int64) *Entry {
	return &Entry{
		id:       id,
		parentID: parentID,
		name:     name,
		childID:  childID,
	}
}

// Getters
func (e *Entry) ID() int64       { return e.id }
func (e *Entry) ParentID() int64 { return e.parentID }
func (e *Entry) Name() string    { return e.name }
func (e *Entry) ChildID() int64  { return e.childID }
