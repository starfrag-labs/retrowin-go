package fs

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode/content"
	inodeMocks "github.com/starfrag-lab/retrowin-go/internal/core/inode/mocks"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// --- GetRootDirectory ---

func TestGetRootDirectory_Success(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, nil, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	result, err := svc.GetRootDirectory(context.Background(), "sys")

	assert.NoError(t, err)
	assert.Equal(t, rootInode, result)
}

func TestGetRootDirectory_NotFound(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)

	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{}, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	_, err := svc.GetRootDirectory(context.Background(), "sys")

	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}

func TestGetRootDirectory_FindFailure(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)

	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return(nil, errors.Internal("find failed"))

	svc := NewService(inodeSvc, nil, nil, nil)

	_, err := svc.GetRootDirectory(context.Background(), "sys")

	assert.Error(t, err)
}

func TestGetRootDirectory_NotDirectory(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	rootInode := inode.NewInode("root-id", "sys", inode.ModeRegular|0644, 0, 0, 0, 1, inode.FlagRoot, now, now, now, nil, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	_, err := svc.GetRootDirectory(context.Background(), "sys")

	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}

// --- ResolvePath ---

func TestResolvePath_Root(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, nil, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	result, err := svc.ResolvePath(context.Background(), "sys", "/")

	assert.NoError(t, err)
	assert.Equal(t, rootInode, result)
}

func TestResolvePath_SingleLevel(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	dirContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "file", InodeID: "file-id", FileType: uint8(inode.ModeRegular >> 12)},
	}}
	raw, _ := json.Marshal(dirContent)
	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, raw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	fileInode := inode.NewInode("file-id", "sys", inode.ModeRegular|0644, 1000, 1000, 0, 1, 0, now, now, now, nil, now, now)
	inodeSvc.EXPECT().GetByID(mock.Anything, "file-id").Return(fileInode, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	result, err := svc.ResolvePath(context.Background(), "sys", "/file")

	assert.NoError(t, err)
	assert.Equal(t, fileInode, result)
}

func TestResolvePath_MultiLevel(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	aContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "b", InodeID: "b-id", FileType: uint8(inode.ModeDirectory >> 12)},
	}}
	aRaw, _ := json.Marshal(aContent)
	rootContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "a", InodeID: "a-id", FileType: uint8(inode.ModeDirectory >> 12)},
	}}
	rootRaw, _ := json.Marshal(rootContent)
	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, rootRaw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	aInode := inode.NewInode("a-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, 0, now, now, now, aRaw, now, now)
	inodeSvc.EXPECT().GetByID(mock.Anything, "a-id").Return(aInode, nil)

	bContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "c", InodeID: "c-id", FileType: uint8(inode.ModeRegular >> 12)},
	}}
	bRaw, _ := json.Marshal(bContent)
	bInode := inode.NewInode("b-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, 0, now, now, now, bRaw, now, now)
	inodeSvc.EXPECT().GetByID(mock.Anything, "b-id").Return(bInode, nil)

	cInode := inode.NewInode("c-id", "sys", inode.ModeRegular|0644, 1000, 1000, 0, 1, 0, now, now, now, nil, now, now)
	inodeSvc.EXPECT().GetByID(mock.Anything, "c-id").Return(cInode, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	result, err := svc.ResolvePath(context.Background(), "sys", "/a/b/c")

	assert.NoError(t, err)
	assert.Equal(t, cInode, result)
}

func TestResolvePath_RelativePath(t *testing.T) {
	svc := NewService(nil, nil, nil, nil)

	_, err := svc.ResolvePath(context.Background(), "sys", "relative/path")

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
}

func TestResolvePath_EmptyPath(t *testing.T) {
	svc := NewService(nil, nil, nil, nil)

	_, err := svc.ResolvePath(context.Background(), "sys", "")

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
}

func TestResolvePath_ComponentNotFound(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	dirContent := content.DirContent{Entries: []content.DirEntry{}}
	raw, _ := json.Marshal(dirContent)
	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, raw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	_, err := svc.ResolvePath(context.Background(), "sys", "/nonexistent")

	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}

func TestResolvePath_NotADirectory(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	dirContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "file", InodeID: "file-id", FileType: uint8(inode.ModeRegular >> 12)},
	}}
	raw, _ := json.Marshal(dirContent)
	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, raw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	fileInode := inode.NewInode("file-id", "sys", inode.ModeRegular|0644, 1000, 1000, 0, 1, 0, now, now, now, nil, now, now)
	inodeSvc.EXPECT().GetByID(mock.Anything, "file-id").Return(fileInode, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	_, err := svc.ResolvePath(context.Background(), "sys", "/file/subpath")

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
}

func TestResolvePath_Symlink(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	symContent := content.SymlinkContent{Target: "/target"}
	symRaw, _ := json.Marshal(symContent)
	dirContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "link", InodeID: "link-id", FileType: uint8(inode.ModeSymlink >> 12)},
	}}
	raw, _ := json.Marshal(dirContent)
	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, raw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil).Once()

	linkInode := inode.NewInode("link-id", "sys", inode.ModeSymlink|0777, 0, 0, 0, 1, 0, now, now, now, symRaw, now, now)
	inodeSvc.EXPECT().GetByID(mock.Anything, "link-id").Return(linkInode, nil)

	targetContent := content.DirContent{Entries: []content.DirEntry{}}
	targetRaw, _ := json.Marshal(targetContent)
	targetInode := inode.NewInode("target-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, 0, now, now, now, targetRaw, now, now)

	targetRootContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "target", InodeID: "target-id", FileType: uint8(inode.ModeDirectory >> 12)},
	}}
	targetRootRaw, _ := json.Marshal(targetRootContent)
	targetRootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, targetRootRaw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{targetRootInode}, nil).Once()

	inodeSvc.EXPECT().GetByID(mock.Anything, "target-id").Return(targetInode, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	result, err := svc.ResolvePath(context.Background(), "sys", "/link")

	assert.NoError(t, err)
	assert.Equal(t, targetInode, result)
}

func TestResolvePath_SymlinkParseError(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	dirContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "link", InodeID: "link-id", FileType: uint8(inode.ModeSymlink >> 12)},
	}}
	raw, _ := json.Marshal(dirContent)
	rootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, raw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{rootInode}, nil)

	linkInode := inode.NewInode("link-id", "sys", inode.ModeSymlink|0777, 0, 0, 0, 1, 0, now, now, now, []byte("invalid"), now, now)
	inodeSvc.EXPECT().GetByID(mock.Anything, "link-id").Return(linkInode, nil)

	svc := NewService(inodeSvc, nil, nil, nil)

	_, err := svc.ResolvePath(context.Background(), "sys", "/link")

	assert.Error(t, err)
}

// --- readDirEntries ---

func TestReadDirEntries_Success(t *testing.T) {
	svc := NewService(nil, nil, nil, nil).(*service)
	now := time.Now()

	entries := []content.DirEntry{
		{Name: "file", InodeID: "file-id", FileType: uint8(inode.ModeRegular >> 12)},
	}
	dirContent := content.DirContent{Entries: entries}
	raw, _ := json.Marshal(dirContent)
	dirInode := inode.NewInode("dir-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, 0, now, now, now, raw, now, now)

	result, err := svc.readDirEntries(dirInode)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "file", result[0].Name)
}

func TestReadDirEntries_NotDirectory(t *testing.T) {
	svc := NewService(nil, nil, nil, nil).(*service)
	now := time.Now()

	fileInode := inode.NewInode("file-id", "sys", inode.ModeRegular|0644, 0, 0, 0, 1, 0, now, now, now, nil, now, now)

	_, err := svc.readDirEntries(fileInode)

	assert.Error(t, err)
	assert.True(t, errors.IsBadRequest(err))
}

func TestReadDirEntries_NilContent(t *testing.T) {
	svc := NewService(nil, nil, nil, nil).(*service)
	now := time.Now()

	dirInode := inode.NewInode("dir-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, 0, now, now, now, nil, now, now)

	result, err := svc.readDirEntries(dirInode)

	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestReadDirEntries_UnparsableContent(t *testing.T) {
	svc := NewService(nil, nil, nil, nil).(*service)
	now := time.Now()

	dirInode := inode.NewInode("dir-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, 0, now, now, now, []byte("invalid"), now, now)

	_, err := svc.readDirEntries(dirInode)

	assert.Error(t, err)
}

// --- resolveSymlink ---

func TestResolveSymlink_Success(t *testing.T) {
	inodeSvc := inodeMocks.NewInodeServiceMock(t)
	now := time.Now()

	symContent := content.SymlinkContent{Target: "/target"}
	symRaw, _ := json.Marshal(symContent)
	symInode := inode.NewInode("sym-id", "sys", inode.ModeSymlink|0777, 0, 0, 0, 1, 0, now, now, now, symRaw, now, now)

	targetContent := content.DirContent{Entries: []content.DirEntry{}}
	targetRaw, _ := json.Marshal(targetContent)
	targetInode := inode.NewInode("target-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, 0, now, now, now, targetRaw, now, now)

	targetRootContent := content.DirContent{Entries: []content.DirEntry{
		{Name: "target", InodeID: "target-id", FileType: uint8(inode.ModeDirectory >> 12)},
	}}
	targetRootRaw, _ := json.Marshal(targetRootContent)
	targetRootInode := inode.NewInode("root-id", "sys", inode.ModeDirectory|0755, 0, 0, 0, 1, inode.FlagRoot, now, now, now, targetRootRaw, now, now)
	inodeSvc.EXPECT().Find(mock.Anything, mock.Anything).Return([]*inode.Inode{targetRootInode}, nil)

	inodeSvc.EXPECT().GetByID(mock.Anything, "target-id").Return(targetInode, nil)

	svc := NewService(inodeSvc, nil, nil, nil).(*service)

	result, err := svc.resolveSymlink(context.Background(), symInode)

	assert.NoError(t, err)
	assert.Equal(t, targetInode, result)
}

func TestResolveSymlink_UnparsableContent(t *testing.T) {
	svc := NewService(nil, nil, nil, nil).(*service)
	now := time.Now()

	symInode := inode.NewInode("sym-id", "sys", inode.ModeSymlink|0777, 0, 0, 0, 1, 0, now, now, now, []byte("invalid"), now, now)

	_, err := svc.resolveSymlink(context.Background(), symInode)

	assert.Error(t, err)
}
