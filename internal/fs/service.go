package fs

import (
	"context"
	"fmt"

	"github.com/starfrag-lab/retrowin-go/internal/directory"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/filedata"
	"github.com/starfrag-lab/retrowin-go/internal/inode"
	"github.com/starfrag-lab/retrowin-go/internal/symlink"
)

// System file types
const (
	SystemTypeRoot  = inode.SystemTypeRoot
	SystemTypeHome  = inode.SystemTypeHome
	SystemTypeTrash = inode.SystemTypeTrash
)

// File type constants
const (
	FileTypeRegular   = inode.FileTypeRegular
	FileTypeDirectory = inode.FileTypeDirectory
	FileTypeSymlink   = inode.FileTypeSymlink
)

// Storage type constants
const (
	StorageTypeS3    = filedata.StorageTypeS3
	StorageTypeLocal = filedata.StorageTypeLocal
)

// CreateFileCommand for creating a complete file.
type CreateFileCommand struct {
	Name       string
	ParentID   *int64
	FileType   inode.FileType
	OwnerUID   string
	OwnerGID   string
	PermOwner  string
	PermGroup  string
	PermOthers string
	IsSystem   bool
	SystemType *string
	// For regular files
	StorageType filedata.StorageType
	Location    string
	Checksum    *string
	// For symlinks
	TargetPath string
}

// MoveCommand for moving a file.
type MoveCommand struct {
	InodeID     int64
	OldParentID int64
	NewParentID int64
	NewName     string
}

// CopyCommand for copying a file.
type CopyCommand struct {
	InodeID        int64
	TargetParentID int64
	NewName        string
}

// RenameCommand for renaming a file.
type RenameCommand struct {
	InodeID  int64
	ParentID int64
	NewName  string
}

// DeleteCommand for deleting a file.
type DeleteCommand struct {
	ParentID int64
	Name     string
}

// UpdateByteSizeCommand for updating file byte size.
type UpdateByteSizeCommand struct {
	InodeID  int64
	ByteSize int64
}

// InitializeUserStorageResult contains the created system directories.
type InitializeUserStorageResult struct {
	Root  *File
	Home  *File
	Trash *File
}

// Service defines the filesystem orchestration interface.
type Service interface {
	// Path-based operations
	GetByPath(ctx context.Context, ownerUID, path string) (*File, error)

	// ID-based operations
	GetByID(ctx context.Context, id int64) (*File, error)
	GetRoot(ctx context.Context, ownerUID string) (*File, error)
	GetHome(ctx context.Context, ownerUID string) (*File, error)
	GetTrash(ctx context.Context, ownerUID string) (*File, error)
	ListDirectory(ctx context.Context, dirInodeID int64) ([]*File, error)

	// Write operations
	CreateFile(ctx context.Context, cmd *CreateFileCommand) (*File, error)
	CreateDirectory(ctx context.Context, parentID int64, name, ownerUID string) (*File, error)
	CreateSymlink(ctx context.Context, parentID int64, name, targetPath, ownerUID string) (*File, error)
	CreateHardLink(ctx context.Context, parentID int64, name string, targetInodeID int64) (*File, error)
	Move(ctx context.Context, cmd *MoveCommand) (*File, error)
	Copy(ctx context.Context, cmd *CopyCommand) (*File, error)
	Delete(ctx context.Context, parentID int64, name string) error
	Rename(ctx context.Context, inodeID, parentID int64, newName string) error

	// Byte size update (for upload completion)
	UpdateByteSize(ctx context.Context, inodeID int64, byteSize int64) error

	// Initialization
	InitializeUserStorage(ctx context.Context, ownerUID string) (*InitializeUserStorageResult, error)
}

type service struct {
	inodeSvc    inode.Service
	dirSvc      directory.Service
	symlinkSvc  symlink.Service
	fileDataSvc filedata.Service
}

// NewService creates a new filesystem Service.
func NewService(
	inodeSvc inode.Service,
	dirSvc directory.Service,
	symlinkSvc symlink.Service,
	fileDataSvc filedata.Service,
) Service {
	return &service{
		inodeSvc:    inodeSvc,
		dirSvc:      dirSvc,
		symlinkSvc:  symlinkSvc,
		fileDataSvc: fileDataSvc,
	}
}

// GetByPath retrieves a file by its full path.
func (s *service) GetByPath(ctx context.Context, ownerUID, path string) (*File, error) {
	root, err := s.GetRoot(ctx, ownerUID)
	if err != nil {
		return nil, err
	}

	components := splitPath(path)
	if len(components) == 0 {
		return root, nil
	}

	current := root.Inode
	for _, name := range components {
		if current.FileType() != inode.FileTypeDirectory {
			return nil, errors.BadRequest("not a directory")
		}

		entry, err := s.dirSvc.FindByParentAndName(ctx, current.ID(), name)
		if err != nil {
			return nil, err
		}
		if entry == nil {
			return nil, errors.NotFound("file not found")
		}

		current, err = s.inodeSvc.GetByID(ctx, entry.ChildID())
		if err != nil {
			return nil, err
		}
	}

	return s.buildFile(ctx, current, nil)
}

// GetByID retrieves a file by inode ID.
func (s *service) GetByID(ctx context.Context, id int64) (*File, error) {
	in, err := s.inodeSvc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	entries, err := s.dirSvc.FindByChild(ctx, id)
	if err != nil {
		return nil, err
	}

	var entry *directory.Entry
	if len(entries) > 0 {
		entry = entries[0]
	}

	return s.buildFile(ctx, in, entry)
}

// GetRoot retrieves the root container for a user.
func (s *service) GetRoot(ctx context.Context, ownerUID string) (*File, error) {
	in, err := s.inodeSvc.FindByOwnerAndSystemType(ctx, ownerUID, SystemTypeRoot)
	if err != nil {
		return nil, err
	}
	return s.buildFile(ctx, in, nil)
}

// GetHome retrieves the home container for a user.
func (s *service) GetHome(ctx context.Context, ownerUID string) (*File, error) {
	in, err := s.inodeSvc.FindByOwnerAndSystemType(ctx, ownerUID, SystemTypeHome)
	if err != nil {
		return nil, err
	}
	return s.buildFile(ctx, in, nil)
}

// GetTrash retrieves the trash container for a user.
func (s *service) GetTrash(ctx context.Context, ownerUID string) (*File, error) {
	in, err := s.inodeSvc.FindByOwnerAndSystemType(ctx, ownerUID, SystemTypeTrash)
	if err != nil {
		return nil, err
	}
	return s.buildFile(ctx, in, nil)
}

// ListDirectory lists all entries in a directory.
func (s *service) ListDirectory(ctx context.Context, dirInodeID int64) ([]*File, error) {
	in, err := s.inodeSvc.GetByID(ctx, dirInodeID)
	if err != nil {
		return nil, err
	}
	if in.FileType() != inode.FileTypeDirectory {
		return nil, errors.BadRequest("not a directory")
	}

	entries, err := s.dirSvc.FindByParent(ctx, dirInodeID)
	if err != nil {
		return nil, err
	}

	files := make([]*File, 0, len(entries))
	for _, entry := range entries {
		childInode, err := s.inodeSvc.GetByID(ctx, entry.ChildID())
		if err != nil || childInode == nil {
			continue
		}
		file, err := s.buildFile(ctx, childInode, entry)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

// CreateFile creates a new regular file.
func (s *service) CreateFile(ctx context.Context, cmd *CreateFileCommand) (*File, error) {
	if cmd.Name == "" {
		return nil, errors.BadRequest("name is required")
	}

	var parentID int64
	if cmd.ParentID != nil {
		parentID = *cmd.ParentID
	} else {
		home, err := s.GetHome(ctx, cmd.OwnerUID)
		if err != nil {
			return nil, err
		}
		parentID = home.ID()
	}

	exists, err := s.dirSvc.Exists(ctx, parentID, cmd.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("file already exists")
	}

	parentInode, err := s.inodeSvc.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if parentInode.FileType() != inode.FileTypeDirectory {
		return nil, errors.BadRequest("not a directory")
	}

	inodeCmd := &inode.CreateCommand{
		FileType:   cmd.FileType,
		OwnerUID:   cmd.OwnerUID,
		OwnerGID:   cmd.OwnerGID,
		PermOwner:  cmd.PermOwner,
		PermGroup:  cmd.PermGroup,
		PermOthers: cmd.PermOthers,
		IsSystem:   cmd.IsSystem,
		SystemType: cmd.SystemType,
	}

	newInode, err := s.inodeSvc.Create(ctx, inodeCmd)
	if err != nil {
		return nil, err
	}

	dirCmd := &directory.CreateCommand{
		ParentID: parentID,
		Name:     cmd.Name,
		ChildID:  newInode.ID(),
	}
	entry, err := s.dirSvc.Create(ctx, dirCmd)
	if err != nil {
		_ = s.inodeSvc.Delete(ctx, newInode.ID())
		return nil, err
	}

	if cmd.StorageType != "" && cmd.Location != "" {
		dataCmd := &filedata.CreateCommand{
			InodeID:     newInode.ID(),
			StorageType: cmd.StorageType,
			Location:    cmd.Location,
			Checksum:    cmd.Checksum,
		}
		_, _ = s.fileDataSvc.Create(ctx, dataCmd)
	}

	return s.buildFile(ctx, newInode, entry)
}

// CreateDirectory creates a new directory.
func (s *service) CreateDirectory(ctx context.Context, parentID int64, name, ownerUID string) (*File, error) {
	if name == "" {
		return nil, errors.BadRequest("name is required")
	}

	exists, err := s.dirSvc.Exists(ctx, parentID, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("file already exists")
	}

	parentInode, err := s.inodeSvc.GetByID(ctx, parentID)
	if err != nil {
		return nil, err
	}
	if parentInode.FileType() != inode.FileTypeDirectory {
		return nil, errors.BadRequest("not a directory")
	}

	inodeCmd := &inode.CreateCommand{
		FileType:   inode.FileTypeDirectory,
		OwnerUID:   ownerUID,
		PermOwner:  "rwx",
		PermGroup:  "r-x",
		PermOthers: "r-x",
	}

	newInode, err := s.inodeSvc.Create(ctx, inodeCmd)
	if err != nil {
		return nil, err
	}

	dirCmd := &directory.CreateCommand{
		ParentID: parentID,
		Name:     name,
		ChildID:  newInode.ID(),
	}
	entry, err := s.dirSvc.Create(ctx, dirCmd)
	if err != nil {
		_ = s.inodeSvc.Delete(ctx, newInode.ID())
		return nil, err
	}

	return s.buildFile(ctx, newInode, entry)
}

// CreateSymlink creates a symbolic link.
func (s *service) CreateSymlink(ctx context.Context, parentID int64, name, targetPath, ownerUID string) (*File, error) {
	if name == "" {
		return nil, errors.BadRequest("name is required")
	}
	if targetPath == "" {
		return nil, errors.BadRequest("target path is required")
	}

	exists, err := s.dirSvc.Exists(ctx, parentID, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("file already exists")
	}

	// Create symlink inode
	inodeCmd := &inode.CreateCommand{
		FileType:   inode.FileTypeSymlink,
		OwnerUID:   ownerUID,
		PermOwner:  "rwx",
		PermGroup:  "r-x",
		PermOthers: "r-x",
	}

	newInode, err := s.inodeSvc.Create(ctx, inodeCmd)
	if err != nil {
		return nil, err
	}

	// Create symlink target
	_, err = s.symlinkSvc.Create(ctx, &symlink.CreateCommand{
		InodeID:    newInode.ID(),
		TargetPath: targetPath,
	})
	if err != nil {
		_ = s.inodeSvc.Delete(ctx, newInode.ID())
		return nil, err
	}

	// Create directory entry
	dirCmd := &directory.CreateCommand{
		ParentID: parentID,
		Name:     name,
		ChildID:  newInode.ID(),
	}
	entry, err := s.dirSvc.Create(ctx, dirCmd)
	if err != nil {
		_ = s.inodeSvc.Delete(ctx, newInode.ID())
		return nil, err
	}

	return s.buildFile(ctx, newInode, entry)
}

// CreateHardLink creates a hard link.
func (s *service) CreateHardLink(ctx context.Context, parentID int64, name string, targetInodeID int64) (*File, error) {
	if name == "" {
		return nil, errors.BadRequest("name is required")
	}

	exists, err := s.dirSvc.Exists(ctx, parentID, name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("file already exists")
	}

	targetInode, err := s.inodeSvc.GetByID(ctx, targetInodeID)
	if err != nil {
		return nil, err
	}
	if targetInode.FileType() != inode.FileTypeRegular {
		return nil, errors.BadRequest("not a regular file")
	}

	dirCmd := &directory.CreateCommand{
		ParentID: parentID,
		Name:     name,
		ChildID:  targetInodeID,
	}
	entry, err := s.dirSvc.Create(ctx, dirCmd)
	if err != nil {
		return nil, err
	}

	// Update link count
	_ = s.inodeSvc.UpdateLinkCount(ctx, targetInodeID, 1)

	return s.buildFile(ctx, targetInode, entry)
}

// Move moves a file to a different directory.
func (s *service) Move(ctx context.Context, cmd *MoveCommand) (*File, error) {
	exists, err := s.dirSvc.Exists(ctx, cmd.NewParentID, cmd.NewName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("file already exists")
	}

	targetInode, err := s.inodeSvc.GetByID(ctx, cmd.NewParentID)
	if err != nil {
		return nil, err
	}
	if targetInode.FileType() != inode.FileTypeDirectory {
		return nil, errors.BadRequest("not a directory")
	}

	if cmd.InodeID == cmd.NewParentID {
		return nil, errors.BadRequest("cannot move directory into itself")
	}

	entries, err := s.dirSvc.FindByChild(ctx, cmd.InodeID)
	if err != nil {
		return nil, err
	}

	var targetEntry *directory.Entry
	for _, e := range entries {
		if e.ParentID() == cmd.OldParentID {
			targetEntry = e
			break
		}
	}

	if targetEntry == nil {
		return nil, errors.NotFound("file not found")
	}

	updateCmd := &directory.UpdateCommand{
		ID:       targetEntry.ID(),
		ParentID: &cmd.NewParentID,
		Name:     &cmd.NewName,
	}
	if err := s.dirSvc.Update(ctx, updateCmd); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, cmd.InodeID)
}

// Copy copies a file (placeholder - requires more complex implementation).
func (s *service) Copy(ctx context.Context, cmd *CopyCommand) (*File, error) {
	return nil, errors.BadRequest("copy operation not yet implemented")
}

// Delete deletes a file.
func (s *service) Delete(ctx context.Context, parentID int64, name string) error {
	entry, err := s.dirSvc.FindByParentAndName(ctx, parentID, name)
	if err != nil {
		return err
	}
	if entry == nil {
		return errors.NotFound("file not found")
	}

	in, err := s.inodeSvc.GetByID(ctx, entry.ChildID())
	if err != nil {
		return err
	}

	if in.IsSystem() {
		return errors.Forbidden("cannot delete root directory")
	}

	if in.FileType() == inode.FileTypeDirectory {
		children, err := s.dirSvc.FindByParent(ctx, in.ID())
		if err != nil {
			return err
		}
		if len(children) > 0 {
			return errors.BadRequest("directory not empty")
		}
	}

	if err := s.dirSvc.Delete(ctx, entry.ID()); err != nil {
		return err
	}

	if in.LinkCount() <= 1 {
		if in.FileType() == inode.FileTypeRegular {
			_ = s.fileDataSvc.Delete(ctx, in.ID())
		}
		if in.FileType() == inode.FileTypeSymlink {
			_ = s.symlinkSvc.Delete(ctx, in.ID())
		}
		return s.inodeSvc.Delete(ctx, in.ID())
	}

	// Decrement link count
	_ = s.inodeSvc.UpdateLinkCount(ctx, in.ID(), -1)

	return nil
}

// Rename renames a file.
func (s *service) Rename(ctx context.Context, inodeID, parentID int64, newName string) error {
	exists, err := s.dirSvc.Exists(ctx, parentID, newName)
	if err != nil {
		return err
	}
	if exists {
		return errors.Conflict("file already exists")
	}

	entries, err := s.dirSvc.FindByChild(ctx, inodeID)
	if err != nil {
		return err
	}

	var targetEntry *directory.Entry
	for _, e := range entries {
		if e.ParentID() == parentID {
			targetEntry = e
			break
		}
	}

	if targetEntry == nil {
		return errors.NotFound("file not found")
	}

	updateCmd := &directory.UpdateCommand{
		ID:   targetEntry.ID(),
		Name: &newName,
	}
	return s.dirSvc.Update(ctx, updateCmd)
}

// UpdateByteSize updates the byte size of a file.
func (s *service) UpdateByteSize(ctx context.Context, inodeID int64, byteSize int64) error {
	cmd := &inode.UpdateCommand{
		ID:       inodeID,
		ByteSize: &byteSize,
	}
	return s.inodeSvc.Update(ctx, cmd)
}

// InitializeUserStorage creates root, home, and trash directories.
func (s *service) InitializeUserStorage(ctx context.Context, ownerUID string) (*InitializeUserStorageResult, error) {
	root, _ := s.GetRoot(ctx, ownerUID)
	if root != nil {
		return nil, nil // Already initialized
	}

	// Create root directory
	rootCmd := &inode.CreateCommand{
		FileType:   inode.FileTypeDirectory,
		OwnerUID:   ownerUID,
		PermOwner:  "rwx",
		PermGroup:  "r-x",
		PermOthers: "r-x",
		IsSystem:   true,
		SystemType: ptrString(SystemTypeRoot),
	}
	rootInode, err := s.inodeSvc.Create(ctx, rootCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create root: %w", err)
	}

	// Create home directory
	homeCmd := &inode.CreateCommand{
		FileType:   inode.FileTypeDirectory,
		OwnerUID:   ownerUID,
		PermOwner:  "rwx",
		PermGroup:  "r-x",
		PermOthers: "r-x",
		IsSystem:   true,
		SystemType: ptrString(SystemTypeHome),
	}
	homeInode, err := s.inodeSvc.Create(ctx, homeCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create home: %w", err)
	}
	_, _ = s.dirSvc.Create(ctx, &directory.CreateCommand{
		ParentID: rootInode.ID(),
		Name:     "home",
		ChildID:  homeInode.ID(),
	})

	// Create trash directory
	trashCmd := &inode.CreateCommand{
		FileType:   inode.FileTypeDirectory,
		OwnerUID:   ownerUID,
		PermOwner:  "rwx",
		PermGroup:  "r-x",
		PermOthers: "r-x",
		IsSystem:   true,
		SystemType: ptrString(SystemTypeTrash),
	}
	trashInode, err := s.inodeSvc.Create(ctx, trashCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create trash: %w", err)
	}
	_, _ = s.dirSvc.Create(ctx, &directory.CreateCommand{
		ParentID: rootInode.ID(),
		Name:     "trash",
		ChildID:  trashInode.ID(),
	})

	return &InitializeUserStorageResult{
		Root:  &File{Inode: rootInode, Path: "/"},
		Home:  &File{Inode: homeInode, Path: "/home"},
		Trash: &File{Inode: trashInode, Path: "/trash"},
	}, nil
}

// buildFile creates a File struct from an Inode and optional DirectoryEntry.
func (s *service) buildFile(ctx context.Context, in *inode.Inode, entry *directory.Entry) (*File, error) {
	file := &File{Inode: in}

	if entry != nil {
		file.Name = entry.Name()
		parentID := entry.ParentID()
		file.ParentID = &parentID
	}

	if in.FileType() == inode.FileTypeRegular {
		fileData, _ := s.fileDataSvc.GetByInodeID(ctx, in.ID())
		file.FileData = fileData
	}

	if in.FileType() == inode.FileTypeSymlink {
		sl, _ := s.symlinkSvc.GetByInodeID(ctx, in.ID())
		file.SymlinkTarget = sl
	}

	path, _ := s.buildPath(ctx, in, entry)
	file.Path = path

	return file, nil
}

// buildPath builds the full path string.
func (s *service) buildPath(ctx context.Context, in *inode.Inode, entry *directory.Entry) (string, error) {
	if entry == nil {
		return "/", nil
	}

	components := []string{entry.Name()}
	currentParentID := entry.ParentID()

	for currentParentID != 0 {
		parentEntries, err := s.dirSvc.FindByChild(ctx, currentParentID)
		if err != nil || len(parentEntries) == 0 {
			break
		}

		parentEntry := parentEntries[0]
		components = append([]string{parentEntry.Name()}, components...)
		currentParentID = parentEntry.ParentID()
	}

	return "/" + joinPath(components), nil
}

// splitPath splits a path into components.
func splitPath(path string) []string {
	if path == "" || path == "/" {
		return nil
	}

	if path[0] == '/' {
		path = path[1:]
	}

	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	if path == "" {
		return nil
	}

	components := make([]string, 0)
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > start {
				components = append(components, path[start:i])
			}
			start = i + 1
		}
	}

	return components
}

// joinPath joins path components.
func joinPath(components []string) string {
	result := ""
	for i, c := range components {
		if i > 0 {
			result += "/"
		}
		result += c
	}
	return result
}

func ptrString(v string) *string {
	return &v
}
