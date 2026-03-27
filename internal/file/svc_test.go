package file_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/starfrag-lab/retrowin-go/internal/file"
	fileMocks "github.com/starfrag-lab/retrowin-go/internal/file/mocks"
)

func TestService_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("returns file when found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		expectedFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "test.txt",
			OwnerID:  100,
		}
		fileRepo.EXPECT().GetByKey(mock.Anything, "test-key").Return(expectedFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1, 2, 3}, nil)

		result, err := svc.Get(ctx, "test-key")

		assert.NoError(t, err)
		assert.Equal(t, expectedFile, result)
		assert.Equal(t, []int64{1, 2, 3}, result.Path)
	})

	t.Run("returns error when file not found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		fileRepo.EXPECT().GetByKey(mock.Anything, "nonexistent").Return(nil, nil)

		result, err := svc.Get(ctx, "nonexistent")

		assert.Error(t, err)
		assert.Equal(t, file.ErrFileNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when repository fails", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		fileRepo.EXPECT().GetByKey(mock.Anything, "test-key").Return(nil, errors.New("db error"))

		result, err := svc.Get(ctx, "test-key")

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns file when found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		expectedFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "test.txt",
			OwnerID:  100,
		}
		fileRepo.EXPECT().GetByID(mock.Anything, int64(1)).Return(expectedFile, nil)

		result, err := svc.GetByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, expectedFile, result)
	})

	t.Run("returns error when file not found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		fileRepo.EXPECT().GetByID(mock.Anything, int64(999)).Return(nil, nil)

		result, err := svc.GetByID(ctx, 999)

		assert.Error(t, err)
		assert.Equal(t, file.ErrFileNotFound, err)
		assert.Nil(t, result)
	})
}

func TestService_GetRoot(t *testing.T) {
	ctx := context.Background()

	t.Run("returns root container when found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		rootFile := &file.File{
			ID:         1,
			FileKey:    "root-key",
			Type:       file.FileTypeContainer,
			FileName:   "Root",
			OwnerID:    100,
			IsSystem:   true,
			SystemType: strPtr("root"),
		}
		fileRepo.EXPECT().GetByOwnerAndSystemType(mock.Anything, int64(100), "root").Return(rootFile, nil)

		result, err := svc.GetRoot(ctx, 100)

		assert.NoError(t, err)
		assert.Equal(t, rootFile, result)
	})

	t.Run("returns error when root not found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		fileRepo.EXPECT().GetByOwnerAndSystemType(mock.Anything, int64(100), "root").Return(nil, nil)

		result, err := svc.GetRoot(ctx, 100)

		assert.Error(t, err)
		assert.Equal(t, file.ErrFileNotFound, err)
		assert.Nil(t, result)
	})
}

func TestService_GetHome(t *testing.T) {
	ctx := context.Background()

	t.Run("returns home container when found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		homeFile := &file.File{
			ID:         2,
			FileKey:    "home-key",
			Type:       file.FileTypeContainer,
			FileName:   "Home",
			OwnerID:    100,
			IsSystem:   true,
			SystemType: strPtr("home"),
		}
		fileRepo.EXPECT().GetByOwnerAndSystemType(mock.Anything, int64(100), "home").Return(homeFile, nil)

		result, err := svc.GetHome(ctx, 100)

		assert.NoError(t, err)
		assert.Equal(t, homeFile, result)
	})
}

func TestService_GetTrash(t *testing.T) {
	ctx := context.Background()

	t.Run("returns trash container when found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		trashFile := &file.File{
			ID:         3,
			FileKey:    "trash-key",
			Type:       file.FileTypeContainer,
			FileName:   "Trash",
			OwnerID:    100,
			IsSystem:   true,
			SystemType: strPtr("trash"),
		}
		fileRepo.EXPECT().GetByOwnerAndSystemType(mock.Anything, int64(100), "trash").Return(trashFile, nil)

		result, err := svc.GetTrash(ctx, 100)

		assert.NoError(t, err)
		assert.Equal(t, trashFile, result)
	})

	t.Run("returns ErrTrashNotFound when trash not found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		fileRepo.EXPECT().GetByOwnerAndSystemType(mock.Anything, int64(100), "trash").Return(nil, nil)

		result, err := svc.GetTrash(ctx, 100)

		assert.Error(t, err)
		assert.Equal(t, file.ErrTrashNotFound, err)
		assert.Nil(t, result)
	})
}

func TestService_GetChildren(t *testing.T) {
	ctx := context.Background()

	t.Run("returns children when parent is container", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		parent := &file.File{
			ID:       1,
			FileKey:  "parent-key",
			Type:     file.FileTypeContainer,
			FileName: "parent",
			OwnerID:  100,
		}
		children := []*file.File{
			{ID: 2, FileKey: "child1", Type: file.FileTypeFile, FileName: "child1.txt", OwnerID: 100},
			{ID: 3, FileKey: "child2", Type: file.FileTypeFile, FileName: "child2.txt", OwnerID: 100},
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "parent-key").Return(parent, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().GetChildren(mock.Anything, int64(1)).Return(children, nil)

		result, err := svc.GetChildren(ctx, "parent-key")

		assert.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("returns error when parent is not a container", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		parent := &file.File{
			ID:       1,
			FileKey:  "parent-key",
			Type:     file.FileTypeFile,
			FileName: "parent.txt",
			OwnerID:  100,
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "parent-key").Return(parent, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)

		result, err := svc.GetChildren(ctx, "parent-key")

		assert.Error(t, err)
		assert.Equal(t, file.ErrNotContainer, err)
		assert.Nil(t, result)
	})
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("creates file successfully without parent", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		cmd := &file.CreateCommand{
			Type:     file.FileTypeFile,
			FileName: "new-file.txt",
			OwnerID:  100,
		}
		newFile := &file.File{
			ID:       1,
			FileKey:  "new-key",
			Type:     file.FileTypeFile,
			FileName: "new-file.txt",
			OwnerID:  100,
		}

		fileRepo.EXPECT().Create(mock.Anything, cmd).Return(newFile, nil)
		infoRepo.EXPECT().Create(mock.Anything, int64(1), int64(0)).Return(&file.FileInfo{FileID: 1}, nil)
		pathRepo.EXPECT().Create(mock.Anything, int64(1), []int64{int64(1)}).Return(nil)
		roleRepo.EXPECT().Create(mock.Anything, int64(100), int64(1), []string{"owner", "read", "write"}).Return(nil)

		result, err := svc.Create(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, newFile, result)
		assert.Equal(t, []int64{1}, result.Path)
		assert.Equal(t, []string{"owner", "read", "write"}, result.Roles)
	})

	t.Run("creates file successfully with parent", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		parentKey := "parent-key"
		cmd := &file.CreateCommand{
			Type:      file.FileTypeFile,
			FileName:  "new-file.txt",
			ParentKey: &parentKey,
			OwnerID:   100,
		}
		parent := &file.File{
			ID:       1,
			FileKey:  "parent-key",
			Type:     file.FileTypeContainer,
			FileName: "parent",
			OwnerID:  100,
		}
		newFile := &file.File{
			ID:       2,
			FileKey:  "new-key",
			Type:     file.FileTypeFile,
			FileName: "new-file.txt",
			OwnerID:  100,
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "parent-key").Return(parent, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().Create(mock.Anything, cmd).Return(newFile, nil)
		infoRepo.EXPECT().Create(mock.Anything, int64(2), int64(0)).Return(&file.FileInfo{FileID: 2}, nil)
		pathRepo.EXPECT().Create(mock.Anything, int64(2), []int64{int64(1), int64(2)}).Return(nil)
		roleRepo.EXPECT().Create(mock.Anything, int64(100), int64(2), []string{"owner", "read", "write"}).Return(nil)

		result, err := svc.Create(ctx, cmd)

		assert.NoError(t, err)
		assert.Equal(t, newFile, result)
		assert.Equal(t, []int64{1, 2}, result.Path)
	})

	t.Run("returns error when file name is empty", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		cmd := &file.CreateCommand{
			Type:     file.FileTypeFile,
			FileName: "",
			OwnerID:  100,
		}

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, "file name is required", err.Error())
		assert.Nil(t, result)
	})

	t.Run("returns error when file type is invalid", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		cmd := &file.CreateCommand{
			Type:     "invalid",
			FileName: "test.txt",
			OwnerID:  100,
		}

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, "invalid file type", err.Error())
		assert.Nil(t, result)
	})

	t.Run("returns error when parent not found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		parentKey := "nonexistent-parent"
		cmd := &file.CreateCommand{
			Type:      file.FileTypeFile,
			FileName:  "test.txt",
			ParentKey: &parentKey,
			OwnerID:   100,
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "nonexistent-parent").Return(nil, nil)

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, file.ErrParentNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when parent is not a container", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		parentKey := "parent-file"
		cmd := &file.CreateCommand{
			Type:      file.FileTypeFile,
			FileName:  "test.txt",
			ParentKey: &parentKey,
			OwnerID:   100,
		}
		parent := &file.File{
			ID:       1,
			FileKey:  "parent-file",
			Type:     file.FileTypeFile,
			FileName: "parent.txt",
			OwnerID:  100,
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "parent-file").Return(parent, nil)

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, file.ErrNotContainer, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when parent owner differs", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		parentKey := "parent-container"
		cmd := &file.CreateCommand{
			Type:      file.FileTypeFile,
			FileName:  "test.txt",
			ParentKey: &parentKey,
			OwnerID:   100,
		}
		parent := &file.File{
			ID:       1,
			FileKey:  "parent-container",
			Type:     file.FileTypeContainer,
			FileName: "parent",
			OwnerID:  200, // Different owner
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "parent-container").Return(parent, nil)

		result, err := svc.Create(ctx, cmd)

		assert.Error(t, err)
		assert.Equal(t, file.ErrAccessDenied, err)
		assert.Nil(t, result)
	})
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("updates file successfully", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "old-name.txt",
			OwnerID:  100,
		}
		newName := "new-name.txt"
		cmd := &file.UpdateCommand{
			FileName: &newName,
		}
		updatedFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "new-name.txt",
			OwnerID:  100,
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "test-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().Update(mock.Anything, int64(1), cmd).Return(updatedFile, nil)

		result, err := svc.Update(ctx, "test-key", cmd)

		assert.NoError(t, err)
		assert.Equal(t, "new-name.txt", result.FileName)
	})

	t.Run("updates byte size and file info", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
		}
		byteSize := int64(2048)
		cmd := &file.UpdateCommand{
			ByteSize: &byteSize,
		}
		updatedFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "test-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().Update(mock.Anything, int64(1), cmd).Return(updatedFile, nil)
		infoRepo.EXPECT().Update(mock.Anything, int64(1), int64(2048)).Return(&file.FileInfo{FileID: 1, ByteSize: 2048}, nil)

		result, err := svc.Update(ctx, "test-key", cmd)

		assert.NoError(t, err)
		assert.Equal(t, int64(2048), result.ByteSize)
	})

	t.Run("returns same file when no updates", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
		}
		cmd := &file.UpdateCommand{} // No updates

		fileRepo.EXPECT().GetByKey(mock.Anything, "test-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)

		result, err := svc.Update(ctx, "test-key", cmd)

		assert.NoError(t, err)
		assert.Equal(t, existingFile, result)
	})
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("moves file to trash on soft delete", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
			IsSystem: false,
		}
		trash := &file.File{
			ID:         2,
			FileKey:    "trash-key",
			Type:       file.FileTypeContainer,
			FileName:   "Trash",
			OwnerID:    100,
			IsSystem:   true,
			SystemType: strPtr("trash"),
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "test-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().GetByOwnerAndSystemType(mock.Anything, int64(100), "trash").Return(trash, nil)
		fileRepo.EXPECT().Move(mock.Anything, int64(1), int64(2)).Return(nil)

		err := svc.Delete(ctx, "test-key", false)

		assert.NoError(t, err)
	})

	t.Run("permanently deletes file", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "test-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
			IsSystem: false,
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "test-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		infoRepo.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)
		pathRepo.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)
		roleRepo.EXPECT().DeleteByFile(mock.Anything, int64(1)).Return(nil)
		fileRepo.EXPECT().Delete(mock.Anything, int64(1)).Return(nil)

		err := svc.Delete(ctx, "test-key", true)

		assert.NoError(t, err)
	})

	t.Run("returns error when trying to delete system file", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		systemFile := &file.File{
			ID:         1,
			FileKey:    "root-key",
			Type:       file.FileTypeContainer,
			FileName:   "Root",
			OwnerID:    100,
			IsSystem:   true,
			SystemType: strPtr("root"),
		}

		fileRepo.EXPECT().GetByKey(mock.Anything, "root-key").Return(systemFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)

		err := svc.Delete(ctx, "root-key", false)

		assert.Error(t, err)
		assert.Equal(t, file.ErrCannotDeleteSystem, err)
	})
}

func TestService_Move(t *testing.T) {
	ctx := context.Background()

	t.Run("moves file successfully", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "file-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
		}
		target := &file.File{
			ID:       2,
			FileKey:  "target-key",
			Type:     file.FileTypeContainer,
			FileName: "target",
			OwnerID:  100,
		}
		cmd := &file.MoveCommand{TargetKey: "target-key"}

		fileRepo.EXPECT().GetByKey(mock.Anything, "file-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().GetByKey(mock.Anything, "target-key").Return(target, nil)
		fileRepo.EXPECT().Move(mock.Anything, int64(1), int64(2)).Return(nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(2)).Return([]int64{2}, nil)
		pathRepo.EXPECT().Update(mock.Anything, int64(1), []int64{int64(2), int64(1)}).Return(nil)

		result, err := svc.Move(ctx, "file-key", cmd)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), *result.ParentID)
	})

	t.Run("returns error when target not found", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "file-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
		}
		cmd := &file.MoveCommand{TargetKey: "nonexistent"}

		fileRepo.EXPECT().GetByKey(mock.Anything, "file-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().GetByKey(mock.Anything, "nonexistent").Return(nil, nil)

		result, err := svc.Move(ctx, "file-key", cmd)

		assert.Error(t, err)
		assert.Equal(t, file.ErrTargetNotFound, err)
		assert.Nil(t, result)
	})

	t.Run("returns error when moving into itself", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "file-key",
			Type:     file.FileTypeContainer,
			FileName: "folder",
			OwnerID:  100,
		}
		cmd := &file.MoveCommand{TargetKey: "file-key"}

		fileRepo.EXPECT().GetByKey(mock.Anything, "file-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().GetByKey(mock.Anything, "file-key").Return(existingFile, nil)

		result, err := svc.Move(ctx, "file-key", cmd)

		assert.Error(t, err)
		assert.Equal(t, file.ErrCannotMoveIntoSelf, err)
		assert.Nil(t, result)
	})
}

func TestService_Copy(t *testing.T) {
	ctx := context.Background()

	t.Run("copies file successfully", func(t *testing.T) {
		fileRepo := fileMocks.NewRepositoryMock(t)
		infoRepo := fileMocks.NewFileInfoRepositoryMock(t)
		pathRepo := fileMocks.NewFilePathRepositoryMock(t)
		roleRepo := fileMocks.NewFileRoleRepositoryMock(t)
		svc := file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)

		existingFile := &file.File{
			ID:       1,
			FileKey:  "file-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
			ByteSize: 1024,
		}
		target := &file.File{
			ID:       2,
			FileKey:  "target-key",
			Type:     file.FileTypeContainer,
			FileName: "target",
			OwnerID:  100,
		}
		newFile := &file.File{
			ID:       3,
			FileKey:  "new-key",
			Type:     file.FileTypeFile,
			FileName: "file.txt",
			OwnerID:  100,
			ByteSize: 1024,
		}
		cmd := &file.CopyCommand{TargetKey: "target-key"}

		fileRepo.EXPECT().GetByKey(mock.Anything, "file-key").Return(existingFile, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(1)).Return([]int64{1}, nil)
		fileRepo.EXPECT().GetByKey(mock.Anything, "target-key").Return(target, nil)
		fileRepo.EXPECT().Copy(mock.Anything, int64(1), int64(2), int64(100)).Return(newFile, nil)
		infoRepo.EXPECT().Create(mock.Anything, int64(3), int64(1024)).Return(&file.FileInfo{FileID: 3}, nil)
		pathRepo.EXPECT().GetByFileID(mock.Anything, int64(2)).Return([]int64{2}, nil)
		pathRepo.EXPECT().Create(mock.Anything, int64(3), []int64{int64(2), int64(3)}).Return(nil)
		roleRepo.EXPECT().GetByUserAndFile(mock.Anything, int64(100), int64(1)).Return([]string{"owner", "read", "write"}, nil)
		roleRepo.EXPECT().Create(mock.Anything, int64(100), int64(3), []string{"owner", "read", "write"}).Return(nil)

		result, err := svc.Copy(ctx, "file-key", cmd)

		assert.NoError(t, err)
		assert.Equal(t, newFile, result)
	})
}

func TestEnsureFileKey(t *testing.T) {
	key1 := file.EnsureFileKey()
	key2 := file.EnsureFileKey()

	// Keys should be valid UUIDs
	assert.Len(t, key1, 36)
	assert.Len(t, key2, 36)

	// Keys should be unique
	assert.NotEqual(t, key1, key2)
}

// Helper function
func strPtr(s string) *string {
	return &s
}
