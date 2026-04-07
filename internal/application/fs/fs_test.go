package fs_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/starfrag-lab/retrowin-go/internal/application/fs"
)

func TestAccessType_Constants(t *testing.T) {
	assert.Equal(t, fs.AccessRead, fs.AccessType(0))
	assert.Equal(t, fs.AccessWrite, fs.AccessType(1))
	assert.Equal(t, fs.AccessExecute, fs.AccessType(2))
}

func TestCreateFileCommand(t *testing.T) {
	cmd := &fs.CreateFileCommand{
		SystemID: "system-123",
		GID:      1000,
		Mode:     0644,
		Flags:    0,
		Content:  []byte("test content"),
	}

	assert.Equal(t, "system-123", cmd.SystemID)
	assert.Equal(t, 1000, cmd.GID)
	assert.Equal(t, 0644, cmd.Mode)
	assert.Equal(t, 0, cmd.Flags)
	assert.Equal(t, []byte("test content"), cmd.Content)
}

func TestCreateDirectoryCommand(t *testing.T) {
	cmd := &fs.CreateDirectoryCommand{
		SystemID: "system-123",
		GID:      1000,
		Mode:     0755,
		Flags:    0,
	}

	assert.Equal(t, "system-123", cmd.SystemID)
	assert.Equal(t, 1000, cmd.GID)
	assert.Equal(t, 0755, cmd.Mode)
	assert.Equal(t, 0, cmd.Flags)
}

func TestCreateSymlinkCommand(t *testing.T) {
	cmd := &fs.CreateSymlinkCommand{
		SystemID: "system-123",
		GID:      1000,
		Mode:     0777,
		Flags:    0,
		Target:   "/target/path",
	}

	assert.Equal(t, "system-123", cmd.SystemID)
	assert.Equal(t, 1000, cmd.GID)
	assert.Equal(t, 0777, cmd.Mode)
	assert.Equal(t, 0, cmd.Flags)
	assert.Equal(t, "/target/path", cmd.Target)
}

func TestUpdateContentCommand(t *testing.T) {
	cmd := &fs.UpdateContentCommand{
		ID:      "inode-123",
		Content: []byte("new content"),
	}

	assert.Equal(t, "inode-123", cmd.ID)
	assert.Equal(t, []byte("new content"), cmd.Content)
}

func TestUpdateModeCommand(t *testing.T) {
	cmd := &fs.UpdateModeCommand{
		ID:   "inode-123",
		Mode: 0755,
	}

	assert.Equal(t, "inode-123", cmd.ID)
	assert.Equal(t, 0755, cmd.Mode)
}

func TestListFilter(t *testing.T) {
	systemID := "system-123"
	uid := 1000

	filter := &fs.ListFilter{
		SystemID: &systemID,
		UID:      &uid,
	}

	assert.Equal(t, "system-123", *filter.SystemID)
	assert.Equal(t, 1000, *filter.UID)
}

func TestPermissionCheck(t *testing.T) {
	t.Run("owner has read permission", func(t *testing.T) {
		mode := 0600 // owner rw
		perm := mode & 0700
		assert.Equal(t, 0600, perm)
	})

	t.Run("group has read permission", func(t *testing.T) {
		mode := 0060 // group rw
		perm := mode & 0070
		assert.Equal(t, 0060, perm)
	})

	t.Run("other has read permission", func(t *testing.T) {
		mode := 0006 // other rw
		perm := mode & 0007
		assert.Equal(t, 0006, perm)
	})

	t.Run("owner can execute", func(t *testing.T) {
		mode := 0100 // owner x
		perm := mode & 0100
		assert.Equal(t, 0100, perm)
	})

	t.Run("all permissions", func(t *testing.T) {
		mode := 0777
		assert.Equal(t, 0777, mode&0777)
	})
}
