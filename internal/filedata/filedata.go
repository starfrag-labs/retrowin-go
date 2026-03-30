package filedata

// StorageType represents the storage backend type.
type StorageType string

const (
	StorageTypeS3    StorageType = "s3"
	StorageTypeLocal StorageType = "local"
)

// FileData represents the storage info for a regular file.
// This is an internal detail, not exposed as a service.
type FileData struct {
	inodeID     int64
	storageType StorageType
	location    string
	checksum    *string
}

// NewFileData creates a new FileData.
func NewFileData(
	inodeID int64,
	storageType StorageType,
	location string,
	checksum *string,
) *FileData {
	return &FileData{
		inodeID:     inodeID,
		storageType: storageType,
		location:    location,
		checksum:    checksum,
	}
}

// Getters
func (f *FileData) InodeID() int64           { return f.inodeID }
func (f *FileData) StorageType() StorageType { return f.storageType }
func (f *FileData) Location() string         { return f.location }
func (f *FileData) Checksum() *string        { return f.checksum }
