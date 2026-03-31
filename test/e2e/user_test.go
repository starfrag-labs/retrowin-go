package e2e

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUser_Get tests the user info endpoint
func TestUser_Get(t *testing.T) {
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

	t.Run("returns current user info", func(t *testing.T) {
		// Create and login user
		user, err := suite.SetupAuthenticatedUser(ctx, "testuser")
		require.NoError(t, err, "Failed to setup authenticated user")

		resp, err := suite.Get("/v1/user")
		require.NoError(t, err, "Failed to get user info")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		assert.Equal(t, user.ID, result["id"], "User ID should match")
		assert.Equal(t, "testuser", result["username"], "Username should match")
	})

	t.Run("returns 401 without session", func(t *testing.T) {
		suite.ClearCookies()

		resp, err := suite.Get("/user")
		require.NoError(t, err, "Failed to make request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"Expected 401 Unauthorized, got %d", resp.StatusCode)
	})
}

// TestSystemUser_Create tests creating a system user
func TestSystemUser_Create(t *testing.T) {
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
	user, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	t.Run("creates user with auto-assigned UID", func(t *testing.T) {
		req := map[string]interface{}{
			"user_id":  user.ID,
			"username": "newuser",
		}

		resp, err := suite.Post("/systems/"+system.ID+"/users", req)
		require.NoError(t, err, "Failed to create system user")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Expected 201 Created, got %d: %s", resp.StatusCode, suite.ReadBody(resp))

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		// UID should be auto-assigned (starting from 1000)
		uid, ok := result["uid"].(float64)
		require.True(t, ok, "UID should be a number")
		assert.GreaterOrEqual(t, int(uid), 1000, "Auto-assigned UID should start from 1000")

		// GID should equal UID (private group)
		gid, ok := result["gid"].(float64)
		require.True(t, ok, "GID should be a number")
		assert.Equal(t, uid, gid, "GID should equal UID for private group")
	})

	t.Run("creates user with explicit UID", func(t *testing.T) {
		req := map[string]interface{}{
			"user_id":  user.ID + "-explicit",
			"username": "explicituser",
			"uid":      2000,
		}

		resp, err := suite.Post("/systems/"+system.ID+"/users", req)
		require.NoError(t, err, "Failed to create system user")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Expected 201 Created, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		uid, ok := result["uid"].(float64)
		require.True(t, ok, "UID should be a number")
		assert.Equal(t, 2000, int(uid), "UID should be the specified value")
	})

	t.Run("rejects duplicate user in same system", func(t *testing.T) {
		// First create should succeed
		req := map[string]interface{}{
			"user_id":  user.ID + "-dup",
			"username": "dupuser",
		}
		resp1, err := suite.Post("/systems/"+system.ID+"/users", req)
		require.NoError(t, err)
		_ = resp1.Body.Close()
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Second create with same user_id should fail
		resp2, err := suite.Post("/systems/"+system.ID+"/users", req)
		require.NoError(t, err)
		defer func() { _ = resp2.Body.Close() }()

		assert.Equal(t, http.StatusConflict, resp2.StatusCode,
			"Expected 409 Conflict for duplicate user, got %d", resp2.StatusCode)
	})

	t.Run("rejects duplicate username in same system", func(t *testing.T) {
		req1 := map[string]interface{}{
			"user_id":  user.ID + "-dupname1",
			"username": "dupusername",
		}
		resp1, err := suite.Post("/systems/"+system.ID+"/users", req1)
		require.NoError(t, err)
		_ = resp1.Body.Close()
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Same username, different user_id
		req2 := map[string]interface{}{
			"user_id":  user.ID + "-dupname2",
			"username": "dupusername",
		}
		resp2, err := suite.Post("/systems/"+system.ID+"/users", req2)
		require.NoError(t, err)
		defer func() { _ = resp2.Body.Close() }()

		assert.Equal(t, http.StatusConflict, resp2.StatusCode,
			"Expected 409 Conflict for duplicate username, got %d", resp2.StatusCode)
	})
}

// TestSystemUser_List tests listing system users
func TestSystemUser_List(t *testing.T) {
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
	user, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	// Create additional system users
	for i := 0; i < 3; i++ {
		req := map[string]interface{}{
			"user_id":  user.ID + string(rune('a'+i)),
			"username": "listuser" + string(rune('a'+i)),
		}
		resp, err := suite.Post("/systems/"+system.ID+"/users", req)
		require.NoError(t, err)
		_ = resp.Body.Close()
	}

	t.Run("lists all users in system", func(t *testing.T) {
		resp, err := suite.Get("/v1/systems/" + system.ID + "/users")
		require.NoError(t, err, "Failed to list system users")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		// Check users array
		users, ok := result["users"].([]interface{})
		if !ok {
			users, ok = result["data"].([]interface{})
		}
		require.True(t, ok, "Response should contain users array")
		assert.GreaterOrEqual(t, len(users), 3,
			"Should have at least 3 users")
	})
}

// TestSystemUser_Delete tests deleting a system user
func TestSystemUser_Delete(t *testing.T) {
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
	user, system, _, err := suite.SetupFullEnvironment(ctx, "testuser")
	require.NoError(t, err, "Failed to setup full environment")

	// Create a system user to delete
	req := map[string]interface{}{
		"user_id":  user.ID + "-delete",
		"username": "deleteuser",
	}
	createResp, err := suite.Post("/systems/"+system.ID+"/users", req)
	require.NoError(t, err)
	var createResult map[string]interface{}
	_ = suite.ReadJSON(createResp, &createResult)
	uid := int(createResult["uid"].(float64))
	_ = createResp.Body.Close()

	t.Run("deletes user by UID", func(t *testing.T) {
		resp, err := suite.Delete("/systems/" + system.ID + "/users/" + strconv.Itoa(uid))
		require.NoError(t, err, "Failed to delete system user")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNoContent, resp.StatusCode,
			"Expected 204 No Content, got %d", resp.StatusCode)

		// Verify user is deleted
		getResp, err := suite.Get("/systems/" + system.ID + "/users/" + strconv.Itoa(uid))
		require.NoError(t, err)
		_ = getResp.Body.Close()
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode,
			"User should not exist after deletion")
	})

	t.Run("returns 404 for non-existent UID", func(t *testing.T) {
		resp, err := suite.Delete("/systems/" + system.ID + "/users/99999")
		require.NoError(t, err, "Failed to make delete request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode,
			"Expected 404 Not Found, got %d", resp.StatusCode)
	})
}

// TestSystemGroup tests group management
func TestSystemGroup(t *testing.T) {
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

	t.Run("creates group with auto-assigned GID", func(t *testing.T) {
		req := map[string]interface{}{
			"name": "developers",
		}

		resp, err := suite.Post("/systems/"+system.ID+"/groups", req)
		require.NoError(t, err, "Failed to create group")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Expected 201 Created, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		gid, ok := result["gid"].(float64)
		require.True(t, ok, "GID should be a number")
		assert.GreaterOrEqual(t, int(gid), 1000, "Auto-assigned GID should start from 1000")
	})

	t.Run("adds user to group", func(t *testing.T) {
		// First create a group
		grpReq := map[string]interface{}{
			"name": "testgroup",
		}
		grpResp, err := suite.Post("/systems/"+system.ID+"/groups", grpReq)
		require.NoError(t, err)
		var grpResult map[string]interface{}
		_ = suite.ReadJSON(grpResp, &grpResult)
		gid := int(grpResult["gid"].(float64))
		_ = grpResp.Body.Close()

		// Create a system user
		userReq := map[string]interface{}{
			"user_id":  "group-user",
			"username": "groupuser",
		}
		userResp, err := suite.Post("/systems/"+system.ID+"/users", userReq)
		require.NoError(t, err)
		var userResult map[string]interface{}
		_ = suite.ReadJSON(userResp, &userResult)
		uid := int(userResult["uid"].(float64))
		_ = userResp.Body.Close()

		// Add user to group
		addResp, err := suite.Post("/systems/"+system.ID+"/groups/"+strconv.Itoa(gid)+"/members/"+strconv.Itoa(uid), nil)
		require.NoError(t, err, "Failed to add user to group")
		defer func() { _ = addResp.Body.Close() }()

		assert.Equal(t, http.StatusNoContent, addResp.StatusCode,
			"Expected 204 No Content, got %d", addResp.StatusCode)
	})
}
