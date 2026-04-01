package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystem_Init tests system initialization scenarios
func TestSystem_Init(t *testing.T) {
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

	// Setup authenticated user
	_, err = suite.SetupAuthenticatedUser(ctx, "testuser")
	require.NoError(t, err, "Failed to setup authenticated user")

	t.Run("initializes system with root user", func(t *testing.T) {
		// Create system via API
		req := map[string]interface{}{
			"name":        "test-system-1",
			"description": "Test system for e2e tests",
		}

		resp, err := suite.Post("/systems", req)
		require.NoError(t, err, "Failed to create system")
		defer func() { _ = resp.Body.Close() }()

		// Read and verify response
		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		require.Equal(t, http.StatusCreated, resp.StatusCode,
			"Expected 201 Created, got %d", resp.StatusCode)

		// Response is wrapped in "system" object
		system, ok := result["system"].(map[string]interface{})
		require.True(t, ok, "Response should contain system object")

		assert.NotEmpty(t, system["id"], "System ID should be set")
		assert.Equal(t, "test-system-1", system["name"], "System name should match")
		assert.Equal(t, "active", system["status"], "System status should be active")
	})

	t.Run("creates multiple systems independently", func(t *testing.T) {
		// Create first system
		req1 := map[string]interface{}{
			"name":        "system-alpha",
			"description": "First test system",
		}
		resp1, err := suite.Post("/systems", req1)
		require.NoError(t, err)
		defer func() { _ = resp1.Body.Close() }()
		require.Equal(t, http.StatusCreated, resp1.StatusCode)

		var result1 map[string]interface{}
		_ = suite.ReadJSON(resp1, &result1)
		system1, _ := result1["system"].(map[string]interface{})

		// Create second system
		req2 := map[string]interface{}{
			"name":        "system-beta",
			"description": "Second test system",
		}
		resp2, err := suite.Post("/systems", req2)
		require.NoError(t, err)
		defer func() { _ = resp2.Body.Close() }()
		require.Equal(t, http.StatusCreated, resp2.StatusCode,
			"Expected 201 Created, got %d: %s", resp2.StatusCode, suite.ReadBody(resp2))

		var result2 map[string]interface{}
		_ = suite.ReadJSON(resp2, &result2)
		system2, _ := result2["system"].(map[string]interface{})

		// Verify systems have different IDs
		assert.NotEqual(t, system1["id"], system2["id"],
			"Different systems should have different IDs")
	})
}

// TestSystem_Get tests system retrieval
func TestSystem_Get(t *testing.T) {
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

	// Setup authenticated user
	_, err = suite.SetupAuthenticatedUser(ctx, "testuser")
	require.NoError(t, err, "Failed to setup authenticated user")

	// Create a test system first
	createResp, err := suite.Post("/systems", map[string]interface{}{
		"name":        "get-test-system",
		"description": "System for get test",
	})
	require.NoError(t, err)
	defer func() { _ = createResp.Body.Close() }()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var createResult map[string]interface{}
	_ = suite.ReadJSON(createResp, &createResult)
	systemData, _ := createResult["system"].(map[string]interface{})
	systemID := systemData["id"].(string)

	t.Run("returns system by ID", func(t *testing.T) {
		resp, err := suite.Get("/systems/" + systemID)
		require.NoError(t, err, "Failed to get system")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		system, ok := result["system"].(map[string]interface{})
		require.True(t, ok, "Response should contain system object")
		assert.Equal(t, systemID, system["id"], "System ID should match")
		assert.Equal(t, "get-test-system", system["name"], "System name should match")
	})

	t.Run("returns 404 for non-existent system", func(t *testing.T) {
		resp, err := suite.Get("/systems/nonexistent-system-id")
		require.NoError(t, err, "Failed to make request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode,
			"Expected 404 Not Found, got %d", resp.StatusCode)
	})
}

// TestSystem_List tests system listing
func TestSystem_List(t *testing.T) {
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

	// Setup authenticated user
	_, err = suite.SetupAuthenticatedUser(ctx, "testuser")
	require.NoError(t, err, "Failed to setup authenticated user")

	// Create multiple test systems
	for i := 0; i < 3; i++ {
		req := map[string]interface{}{
			"name":        "list-test-system-" + string(rune('a'+i)),
			"description": "System for list test",
		}
		resp, err := suite.Post("/systems", req)
		require.NoError(t, err)
		_ = resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	t.Run("lists all systems", func(t *testing.T) {
		resp, err := suite.Get("/systems")
		require.NoError(t, err, "Failed to list systems")
		defer func() { _ = resp.Body.Close() }()

		require.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var result map[string]interface{}
		err = suite.ReadJSON(resp, &result)
		require.NoError(t, err, "Failed to read response JSON")

		// Check that systems array exists and has at least our created systems
		systems, ok := result["systems"].([]interface{})
		if !ok {
			systems, ok = result["data"].([]interface{})
		}
		require.True(t, ok, "Response should contain systems array")
		assert.GreaterOrEqual(t, len(systems), 3,
			"Should have at least 3 systems")
	})
}

// TestSystem_Unauthorized tests unauthorized access
func TestSystem_Unauthorized(t *testing.T) {
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

	// Don't login - test without authentication
	suite.ClearCookies()

	t.Run("rejects system creation without auth", func(t *testing.T) {
		req := map[string]interface{}{
			"name":        "unauthorized-system",
			"description": "Should fail",
		}

		resp, err := suite.Post("/systems", req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"Expected 401 Unauthorized, got %d", resp.StatusCode)
	})

	t.Run("rejects system list without auth", func(t *testing.T) {
		resp, err := suite.Get("/systems")
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"Expected 401 Unauthorized, got %d", resp.StatusCode)
	})
}
