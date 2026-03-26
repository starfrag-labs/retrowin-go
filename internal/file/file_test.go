package file

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileType_Constants(t *testing.T) {
	assert.Equal(t, FileType("container"), FileTypeContainer)
	assert.Equal(t, FileType("file"), FileTypeFile)
}

func TestSystemType_Constants(t *testing.T) {
	assert.Equal(t, "root", SystemTypeRoot)
	assert.Equal(t, "home", SystemTypeHome)
	assert.Equal(t, "trash", SystemTypeTrash)
}

func TestFile_Struct(t *testing.T) {
	parentID := int64(100)
	f := File{
		ID:         1,
		FileKey:    "test-key",
		Type:       FileTypeFile,
		FileName:   "test.txt",
		OwnerID:    10,
		ParentID:   &parentID,
		ByteSize:   1024,
		IsSystem:   false,
		SystemType: nil,
		Path:       []int64{1, 2, 3},
		Roles:      []string{"owner", "read"},
	}

	assert.Equal(t, int64(1), f.ID)
	assert.Equal(t, "test-key", f.FileKey)
	assert.Equal(t, FileTypeFile, f.Type)
	assert.Equal(t, "test.txt", f.FileName)
	assert.Equal(t, int64(10), f.OwnerID)
	assert.Equal(t, &parentID, f.ParentID)
	assert.Equal(t, int64(1024), f.ByteSize)
	assert.False(t, f.IsSystem)
	assert.Nil(t, f.SystemType)
	assert.Equal(t, []int64{1, 2, 3}, f.Path)
	assert.Equal(t, []string{"owner", "read"}, f.Roles)
}

func TestFileInfo_Struct(t *testing.T) {
	info := FileInfo{
		FileID:   1,
		ByteSize: 2048,
	}

	assert.Equal(t, int64(1), info.FileID)
	assert.Equal(t, int64(2048), info.ByteSize)
}

func TestCreateCommand_Struct(t *testing.T) {
	parentKey := "parent-key"
	cmd := CreateCommand{
		Type:      FileTypeContainer,
		FileName:  "new-folder",
		ParentKey: &parentKey,
		OwnerID:   1,
	}

	assert.Equal(t, FileTypeContainer, cmd.Type)
	assert.Equal(t, "new-folder", cmd.FileName)
	assert.Equal(t, &parentKey, cmd.ParentKey)
	assert.Equal(t, int64(1), cmd.OwnerID)
}

func TestCreateCommand_NilParentKey(t *testing.T) {
	cmd := CreateCommand{
		Type:      FileTypeFile,
		FileName:  "file.txt",
		ParentKey: nil,
		OwnerID:   1,
	}

	assert.Nil(t, cmd.ParentKey)
}

func TestUpdateCommand_Struct(t *testing.T) {
	fileName := "renamed.txt"
	byteSize := int64(4096)
	cmd := UpdateCommand{
		FileName: &fileName,
		ByteSize: &byteSize,
	}

	assert.Equal(t, &fileName, cmd.FileName)
	assert.Equal(t, &byteSize, cmd.ByteSize)
}

func TestMoveCommand_Struct(t *testing.T) {
	cmd := MoveCommand{
		TargetKey: "target-folder-key",
	}

	assert.Equal(t, "target-folder-key", cmd.TargetKey)
}

func TestCopyCommand_Struct(t *testing.T) {
	cmd := CopyCommand{
		TargetKey: "destination-key",
	}

	assert.Equal(t, "destination-key", cmd.TargetKey)
}

func TestErrors_Defined(t *testing.T) {
	// Verify all errors are defined
	assert.Error(t, ErrFileNotFound)
	assert.Error(t, ErrParentNotFound)
	assert.Error(t, ErrNotContainer)
	assert.Error(t, ErrAccessDenied)
	assert.Error(t, ErrTrashNotFound)
	assert.Error(t, ErrTargetNotFound)
	assert.Error(t, ErrCannotDeleteSystem)
	assert.Error(t, ErrCannotMoveIntoSelf)

	// Verify error messages
	assert.Equal(t, "file not found", ErrFileNotFound.Error())
	assert.Equal(t, "parent not found", ErrParentNotFound.Error())
	assert.Equal(t, "file is not a container", ErrNotContainer.Error())
	assert.Equal(t, "access denied", ErrAccessDenied.Error())
	assert.Equal(t, "trash container not found", ErrTrashNotFound.Error())
	assert.Equal(t, "target container not found", ErrTargetNotFound.Error())
	assert.Equal(t, "cannot delete system files", ErrCannotDeleteSystem.Error())
	assert.Equal(t, "cannot move file into itself", ErrCannotMoveIntoSelf.Error())
}
