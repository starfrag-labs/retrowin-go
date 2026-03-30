package content

// ObjectContent is stored in inode content for ModeObject inodes.
// References an Object entity in external storage.
type ObjectContent struct {
	ObjectID string `json:"object_id"`
}
