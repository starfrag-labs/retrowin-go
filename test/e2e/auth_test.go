package e2e

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuth_LoginRedirect tests that login redirects to OIDC provider
// NOTE: This test requires a real OIDC provider, so it's skipped in e2e tests
func TestAuth_LoginRedirect(t *testing.T) {
	t.Skip("Skipping - requires real OIDC provider")

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

	t.Run("redirects to OIDC provider with proper parameters", func(t *testing.T) {
		resp, err := suite.Get("/auth/login")
		require.NoError(t, err, "Failed to make login request")
		defer func() { _ = resp.Body.Close() }()

		// Should redirect (302 or 303)
		assert.True(t, resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusSeeOther,
			"Expected redirect status, got %d", resp.StatusCode)

		location := resp.Header.Get("Location")
		assert.NotEmpty(t, location, "Location header should be set")
		assert.Contains(t, location, "keycloak", "Should redirect to Keycloak")
	})
}

// TestAuth_Logout tests the logout flow
func TestAuth_Logout(t *testing.T) {
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

	t.Run("returns 200 with empty logoutUrl without session", func(t *testing.T) {
		suite.ClearCookies()

		resp, err := suite.Post("/auth/logout", nil)
		require.NoError(t, err, "Failed to make logout request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK, got %d", resp.StatusCode)

		var body map[string]any
		err = suite.ReadJSON(resp, &body)
		require.NoError(t, err)
		assert.Equal(t, "", body["logoutUrl"], "logoutUrl should be empty without session")
	})

	t.Run("deletes session and clears cookie", func(t *testing.T) {
		// Setup: Create user and session
		user, err := suite.CreateTestUser(ctx, "keycloak", "test-user-1", "testuser1")
		require.NoError(t, err, "Failed to create test user")

		err = suite.LoginAs(ctx, user.ID)
		require.NoError(t, err, "Failed to login")

		// Verify we're logged in by accessing a protected endpoint
		resp, err := suite.Get("/user")
		require.NoError(t, err, "Failed to access protected endpoint")
		_ = resp.Body.Close()

		// Now logout
		logoutResp, err := suite.Post("/auth/logout", nil)
		require.NoError(t, err, "Failed to make logout request")
		defer func() { _ = logoutResp.Body.Close() }()

		assert.Equal(t, http.StatusOK, logoutResp.StatusCode,
			"Expected 200 OK, got %d", logoutResp.StatusCode)

		// Verify response contains logoutUrl (empty since test sessions have no id_token)
		var logoutBody map[string]any
		err = suite.ReadJSON(logoutResp, &logoutBody)
		require.NoError(t, err)
		assert.Equal(t, "", logoutBody["logoutUrl"], "logoutUrl should be empty for test sessions")

		// Verify session cookie is cleared
		var sessionCleared bool
		for _, cookie := range logoutResp.Cookies() {
			if cookie.Name == "session_id" && (cookie.MaxAge < 0 || cookie.Value == "") {
				sessionCleared = true
				break
			}
		}
		assert.True(t, sessionCleared, "Session cookie should be cleared")
	})

	t.Run("deletes session and clears cookie with session_id", func(t *testing.T) {
		// Setup: Create user and session
		user, err := suite.CreateTestUser(ctx, "keycloak", "test-user-2", "testuser2")
		require.NoError(t, err, "Failed to create test user")

		err = suite.LoginAs(ctx, user.ID)
		require.NoError(t, err, "Failed to login")

		// Verify we're logged in
		resp, err := suite.Get("/user")
		require.NoError(t, err, "Failed to access protected endpoint")
		_ = resp.Body.Close()

		// Now logout
		logoutResp, err := suite.Post("/auth/logout", nil)
		require.NoError(t, err, "Failed to make logout request")
		defer func() { _ = logoutResp.Body.Close() }()

		assert.Equal(t, http.StatusOK, logoutResp.StatusCode,
			"Expected 200 OK, got %d", logoutResp.StatusCode)

		// Verify session_id cookie is cleared
		var sessionCleared bool
		for _, cookie := range logoutResp.Cookies() {
			if cookie.Name == "session_id" && (cookie.MaxAge < 0 || cookie.Value == "") {
				sessionCleared = true
				break
			}
		}
		assert.True(t, sessionCleared, "session_id cookie should be cleared")
	})
}

// TestAuth_SessionValidation tests session validation
func TestAuth_SessionValidation(t *testing.T) {
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

	t.Run("returns 401 without session for protected endpoint", func(t *testing.T) {
		suite.ClearCookies()

		resp, err := suite.Get("/user")
		require.NoError(t, err, "Failed to make request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"Expected 401 Unauthorized, got %d", resp.StatusCode)
	})

	t.Run("accepts valid session", func(t *testing.T) {
		// Create user and session
		user, err := suite.CreateTestUser(ctx, "keycloak", "test-user-3", "testuser3")
		require.NoError(t, err, "Failed to create test user")

		err = suite.LoginAs(ctx, user.ID)
		require.NoError(t, err, "Failed to login")

		// Access protected endpoint
		resp, err := suite.Get("/user")
		require.NoError(t, err, "Failed to access protected endpoint")
		defer func() { _ = resp.Body.Close() }()

		body := suite.ReadBody(resp)
		t.Logf("Response body: %s", body)
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Expected 200 OK with valid session, got %d", resp.StatusCode)
	})

	t.Run("rejects invalid session ID", func(t *testing.T) {
		suite.ClearCookies()

		// Set invalid session cookie
		suite.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "invalid-session-id",
		})

		resp, err := suite.Get("/user")
		require.NoError(t, err, "Failed to make request")
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
			"Expected 401 Unauthorized with invalid session, got %d", resp.StatusCode)
	})
}
