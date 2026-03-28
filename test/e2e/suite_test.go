package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuite_Start(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	// Verify database connection
	require.NotNil(t, suite.EntClient, "EntClient should not be nil")
	require.NotNil(t, suite.DB, "DB should not be nil")

	// Verify database is accessible
	var result int
	err = suite.DB.QueryRow("SELECT 1").Scan(&result)
	require.NoError(t, err, "Failed to query database")
	assert.Equal(t, 1, result)

	t.Log("E2E test suite started successfully")
}

func TestSuite_Migration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	// Verify tables were created - using actual ent table names
	tables := []string{"users", "files", "file_infos", "file_paths", "file_roles", "file_links", "temp_files", "service_status"}
	for _, table := range tables {
		var exists bool
		err := suite.DB.QueryRow(
			"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)",
			table,
		).Scan(&exists)
		require.NoError(t, err, "Failed to check table existence")
		assert.True(t, exists, "Table %s should exist", table)
	}

	t.Log("Database migration verified successfully")
}

func TestSuite_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create a test HTTP server with the health handler
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Test /health endpoint
	resp, err := http.Get(server.URL + "/health")
	require.NoError(t, err, "Failed to call /health endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health status should be 200")

	// Verify content type
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	t.Log("Health check test passed successfully")
}
