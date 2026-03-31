package e2e

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFs_Stat tests inode stat operations
func TestFs_Stat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	err = suite.StartServer(ctx)
	require.NoError(t, err, "Failed to start server")

	// Setup user and system
	_, system, sysUser, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	t.Run("returns root directory info", func(t *testing.T) {
		resp, err := suite.Get("/fs/" + system.ID + "/root")
		require.NoError(t, err, "Failed to stat root")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		// Check inode field
		inode, ok := result["inode"].(map[string]interface{})
		require.True(t, ok, "Response should contain inode object")

		// Verify root directory attributes
		assert.NotEmpty(t, inode["id"], "Root should have an ID")
		assert.Equal(t, system.ID, inode["system_id"], "System ID should match")

		// Mode should be a directory (040000 = S_IFDIR)
		mode, ok := inode["mode"].(float64)
		require.True(t, ok, "Mode should be a number")
		assert.NotZero(t, int(mode)&040000, "Should be a directory")

		_ = sysUser
	})

	t.Run("returns inode by path", func(t *testing.T) {
		// Use query parameter for path
		resp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/home"))
		require.NoError(t, err, "Failed to stat /home")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d: %s", resp.StatusCode, suite.ReadBody(resp))

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		inode, ok := result["inode"].(map[string]interface{})
		require.True(t, ok, "Response should contain inode object")
		assert.NotEmpty(t, inode["id"], "Inode should have an ID")
	})

	t.Run("returns 404 for non-existent path", func(t *testing.T) {
		resp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/nonexistent/path"))
		require.NoError(t, err, "Failed to make request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode,
			"Expected 404 Not Found, got %d", resp.StatusCode)
	})
}

// TestFs_ReadDir tests directory listing
func TestFs_ReadDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	err = suite.StartServer(ctx)
	require.NoError(t, err, "Failed to start server")

	// Setup user and system
	_, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	t.Run("lists root directory", func(t *testing.T) {
		resp, err := suite.Get("/fs/" + system.ID + "/readdir?path=/")
		require.NoError(t, err, "Failed to read root directory")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		// Check entries array
		entries, ok := result["entries"].([]interface{})
		require.True(t, ok, "Response should contain entries array")

		// Root should have at least home directory
		assert.GreaterOrEqual(t, len(entries), 1,
			"Root should have at least home directory")
	})

	t.Run("lists directory by path", func(t *testing.T) {
		resp, err := suite.Get("/fs/" + system.ID + "/readdir?path=" + url.QueryEscape("/home"))
		require.NoError(t, err, "Failed to read /home")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		// Should have entries
		entries, ok := result["entries"].([]interface{})
		require.True(t, ok, "Response should contain entries array")
		_ = entries
	})

	t.Run("returns 404 for non-existent directory", func(t *testing.T) {
		resp, err := suite.Get("/fs/" + system.ID + "/readdir?path=" + url.QueryEscape("/nonexistent/dir"))
		require.NoError(t, err, "Failed to make request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode,
			"Expected 404 Not Found, got %d", resp.StatusCode)
	})
}

// TestFs_Mkdir tests directory creation
func TestFs_Mkdir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	err = suite.StartServer(ctx)
	require.NoError(t, err, "Failed to start server")

	// Setup user and system
	_, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	t.Run("creates directory with default permissions", func(t *testing.T) {
		req := map[string]interface{}{
			"path": "/home/newdir",
		}

		resp, err := suite.Post("/fs/"+system.ID+"/mkdir", req)
		require.NoError(t, err, "Failed to create directory")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Expected 201 Created, got %d: %s", resp.StatusCode, suite.ReadBody(resp))

		// Verify directory was created
		statResp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/home/newdir"))
		require.NoError(t, err)
		defer func() { _ = statResp.Body.Close() }()
		require.Equal(t, http.StatusOK, statResp.StatusCode)
	})

	t.Run("creates directory with custom permissions", func(t *testing.T) {
		req := map[string]interface{}{
			"path": "/home/privatedir",
			"mode": 0700,
		}

		resp, err := suite.Post("/fs/"+system.ID+"/mkdir", req)
		require.NoError(t, err, "Failed to create directory")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Expected 201 Created, got %d", resp.StatusCode)

		// Verify permissions
		statResp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/home/privatedir"))
		require.NoError(t, err)
		var result map[string]interface{}
		_ = suite.ReadJSON(statResp, &result)
		_ = statResp.Body.Close()

		inode, ok := result["inode"].(map[string]interface{})
		require.True(t, ok, "Response should contain inode")
		mode := int(inode["mode"].(float64))
		assert.Equal(t, 0700, mode&0777, "Permissions should be 0700")
	})

	t.Run("rejects duplicate directory", func(t *testing.T) {
		req := map[string]interface{}{
			"path": "/home/dupdir",
		}

		// First create should succeed
		resp1, err := suite.Post("/fs/"+system.ID+"/mkdir", req)
		require.NoError(t, err)
		_ = resp1.Body.Close()
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Second create should fail
		resp2, err := suite.Post("/fs/"+system.ID+"/mkdir", req)
		require.NoError(t, err)
		defer func() { _ = resp2.Body.Close() }()

		assert.Equal(t, http.StatusConflict, resp2.StatusCode,
			"Expected 409 Conflict for duplicate, got %d", resp2.StatusCode)
	})
}

// TestFs_Delete tests inode deletion
func TestFs_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	err = suite.StartServer(ctx)
	require.NoError(t, err, "Failed to start server")

	// Setup user and system
	_, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	t.Run("deletes empty directory", func(t *testing.T) {
		// Create directory first
		mkdirReq := map[string]interface{}{
			"path": "/home/deletedir",
		}
		mkdirResp, err := suite.Post("/fs/"+system.ID+"/mkdir", mkdirReq)
		require.NoError(t, err)
		_ = mkdirResp.Body.Close()
		require.Equal(t, http.StatusCreated, mkdirResp.StatusCode)

		// Delete the directory
		deleteResp, err := suite.Delete("/fs/" + system.ID + "/unlink?path=" + url.QueryEscape("/home/deletedir"))
		require.NoError(t, err, "Failed to delete directory")
		defer func() { _ = deleteResp.Body.Close() }()

		assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode,
			"Expected 204 No Content, got %d", deleteResp.StatusCode)

		// Verify directory is deleted
		statResp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/home/deletedir"))
		require.NoError(t, err)
		_ = statResp.Body.Close()
		assert.Equal(t, http.StatusNotFound, statResp.StatusCode,
			"Directory should not exist after deletion")
	})

	t.Run("rejects non-empty directory", func(t *testing.T) {
		// Create directory with a subdirectory inside
		mkdirReq := map[string]interface{}{
			"path": "/home/nonemptydir",
		}
		mkdirResp, err := suite.Post("/fs/"+system.ID+"/mkdir", mkdirReq)
		require.NoError(t, err)
		_ = mkdirResp.Body.Close()

		// Create subdirectory
		subReq := map[string]interface{}{
			"path": "/home/nonemptydir/subdir",
		}
		subResp, err := suite.Post("/fs/"+system.ID+"/mkdir", subReq)
		require.NoError(t, err)
		_ = subResp.Body.Close()

		// Try to delete non-empty directory
		deleteResp, err := suite.Delete("/fs/" + system.ID + "/unlink?path=" + url.QueryEscape("/home/nonemptydir"))
		require.NoError(t, err, "Failed to make delete request")
		defer func() { _ = deleteResp.Body.Close() }()

		assert.Equal(t, http.StatusBadRequest, deleteResp.StatusCode,
			"Expected 400 Bad Request for non-empty directory, got %d", deleteResp.StatusCode)
	})

	t.Run("returns 404 for non-existent path", func(t *testing.T) {
		deleteResp, err := suite.Delete("/fs/" + system.ID + "/unlink?path=" + url.QueryEscape("/home/nonexistent"))
		require.NoError(t, err, "Failed to make delete request")
		defer func() { _ = deleteResp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, deleteResp.StatusCode,
			"Expected 404 Not Found, got %d", deleteResp.StatusCode)
	})
}

// TestFs_Symlink tests symbolic link creation
func TestFs_Symlink(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	err = suite.StartServer(ctx)
	require.NoError(t, err, "Failed to start server")

	// Setup user and system
	_, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	// Create a target directory first (symlink target needs to exist for useful test)
	mkdirReq := map[string]interface{}{
		"path": "/home/targetdir",
	}
	mkdirResp, err := suite.Post("/fs/"+system.ID+"/mkdir", mkdirReq)
	require.NoError(t, err)
	_ = mkdirResp.Body.Close()
	require.Equal(t, http.StatusCreated, mkdirResp.StatusCode)

	t.Run("creates symbolic link", func(t *testing.T) {
		req := map[string]interface{}{
			"target":    "/home/targetdir",
			"link_path": "/home/linkdir",
		}

		resp, err := suite.Post("/fs/"+system.ID+"/symlink", req)
		require.NoError(t, err, "Failed to create symlink")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Expected 201 Created, got %d: %s", resp.StatusCode, suite.ReadBody(resp))

		// Verify link was created
		statResp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/home/linkdir"))
		require.NoError(t, err)
		var result map[string]interface{}
		_ = suite.ReadJSON(statResp, &result)
		_ = statResp.Body.Close()

		inode, ok := result["inode"].(map[string]interface{})
		require.True(t, ok, "Response should contain inode")

		// Mode should indicate symlink
		mode := int(inode["mode"].(float64))
		assert.NotZero(t, mode&0120000, "Should be a symlink") // S_IFLNK
	})

	t.Run("can create dangling symlink", func(t *testing.T) {
		req := map[string]interface{}{
			"target":    "/home/nonexistent.txt",
			"link_path": "/home/dangling.txt",
		}

		resp, err := suite.Post("/fs/"+system.ID+"/symlink", req)
		require.NoError(t, err, "Failed to create dangling symlink")
		defer func() { _ = resp.Body.Close() }()

		// Dangling symlinks should still be created successfully
		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Symlink to non-existent target should still be created")
	})
}

// TestFs_Chmod tests permission changes
func TestFs_Chmod(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	err = suite.StartServer(ctx)
	require.NoError(t, err, "Failed to start server")

	// Setup user and system
	_, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	// Create a directory first
	mkdirReq := map[string]interface{}{
		"path": "/home/chmoddir",
		"mode": 0755,
	}
	mkdirResp, err := suite.Post("/fs/"+system.ID+"/mkdir", mkdirReq)
	require.NoError(t, err)
	_ = mkdirResp.Body.Close()
	require.Equal(t, http.StatusCreated, mkdirResp.StatusCode)

	t.Run("changes directory permissions", func(t *testing.T) {
		req := map[string]interface{}{
			"path": "/home/chmoddir",
			"mode": 0700,
		}

		resp, err := suite.Patch("/fs/"+system.ID+"/chmod", req)
		require.NoError(t, err, "Failed to chmod")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		// Verify permissions changed
		statResp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/home/chmoddir"))
		require.NoError(t, err)
		var result map[string]interface{}
		_ = suite.ReadJSON(statResp, &result)
		_ = statResp.Body.Close()

		inode, ok := result["inode"].(map[string]interface{})
		require.True(t, ok, "Response should contain inode")
		mode := int(inode["mode"].(float64))
		assert.Equal(t, 0700, mode&0777, "Permissions should be 0700")
	})

	t.Run("returns 404 for non-existent path", func(t *testing.T) {
		req := map[string]interface{}{
			"path": "/home/nonexistent",
			"mode": 0755,
		}

		resp, err := suite.Patch("/fs/"+system.ID+"/chmod", req)
		require.NoError(t, err, "Failed to make request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode,
			"Expected 404 Not Found, got %d", resp.StatusCode)
	})
}

// TestFs_Permission tests permission enforcement
func TestFs_Permission(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	err = suite.StartServer(ctx)
	require.NoError(t, err, "Failed to start server")

	// Setup user and system
	_, system, sysUser, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	_ = system
	_ = sysUser

	// Create a directory with restricted permissions
	mkdirReq := map[string]interface{}{
		"path": "/home/owneronly",
		"mode": 0700,
	}
	mkdirResp, err := suite.Post("/fs/"+system.ID+"/mkdir", mkdirReq)
	require.NoError(t, err)
	_ = mkdirResp.Body.Close()
	require.Equal(t, http.StatusCreated, mkdirResp.StatusCode)

	t.Run("owner can access owner-only directory", func(t *testing.T) {
		// Owner should be able to stat the directory
		statResp, err := suite.Get("/fs/" + system.ID + "/stat?path=" + url.QueryEscape("/home/owneronly"))
		require.NoError(t, err)
		defer func() { _ = statResp.Body.Close() }()
		require.Equal(t, http.StatusOK, statResp.StatusCode)
	})

	// TODO: Add tests for:
	// - group can read group-readable file
	// - others cannot read owner-only file
	// - root can access any file
	// These require additional user setup
}
