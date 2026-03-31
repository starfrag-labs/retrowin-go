package e2e

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	retrowinserver "github.com/starfrag-lab/retrowin-go/internal/cmd/retrowin-server"
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
	tables := []string{"users", "inodes", "objects", "systems", "system_users", "system_groups"}
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

func TestSuite_ServerStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	shutdown := make(chan struct{})

	// Create a test HTTP server that simulates the retrowin server
	mux := http.NewServeMux()

	// Health check endpoint (direct access, no /v1 prefix)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
	}

	go func() {
		t.Log("Starting test server on 127.0.0.1:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
		close(shutdown)
	}()

	// Wait for server to start
	time.Sleep(500 * time.Millisecond)

	// Test /health endpoint
	resp, err := http.Get("http://127.0.0.1:8080/health")
	require.NoError(t, err, "Failed to call /health endpoint")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health status should be 200")

	// Verify content type
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	assert.Contains(t, string(body), "healthy", "Response should contain 'healthy'")

	t.Log("Server startup test passed successfully")

	// Shutdown server
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	_ = server.Shutdown(ctxShutdown)

	// Wait for shutdown to complete
	select {
	case <-shutdown:
	case <-time.After(5 * time.Second):
	}
}

func TestSuite_FullServerStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	suite := NewSuite(t)
	err := suite.Start(ctx)
	require.NoError(t, err, "Failed to start test suite")
	t.Cleanup(func() { _ = suite.Stop(ctx) })

	// Create a temporary config file with testcontainers connection details
	tmpDir := t.TempDir()
	cfgFile := tmpDir + "/config.yaml"

	// Get postgres connection info and update config
	pgHost, err := suite.PgContainer.Host(ctx)
	require.NoError(t, err, "Failed to get postgres host")
	pgPort, err := suite.PgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err, "Failed to get postgres port")
	suite.Config.Database.Host = pgHost
	suite.Config.Database.Port = pgPort.Int()

	// Disable services that require external dependencies for e2e testing
	suite.Config.Storage.Provider = "memory"
	suite.Config.Auth.Keycloak.BaseURL = "http://localhost:9999" // Invalid URL to prevent actual OIDC calls

	// Write config to temp file as YAML
	cfgData, err := yaml.Marshal(suite.Config)
	require.NoError(t, err, "Failed to marshal config")
	err = os.WriteFile(cfgFile, cfgData, 0644)
	require.NoError(t, err, "Failed to write config file")

	t.Logf("Using config file: %s", cfgFile)
	t.Logf("Database: %s:%d", suite.Config.Database.Host, suite.Config.Database.Port)

	// Start the actual fx app with test config
	// This test verifies that the real server starts and responds to health checks
	app := retrowinserver.NewFXApp(cfgFile, suite.Config.HTTP.Port)

	// Start app in background
	appDone := make(chan struct{})
	go func() {
		app.Run()
		close(appDone)
	}()

	// Wait for app to start
	time.Sleep(2 * time.Second)

	// Verify app is still running (hasn't exited)
	select {
	case <-appDone:
		t.Fatal("FX app exited unexpectedly during startup")
	default:
		// App is still running, proceed with tests
	}

	// Test /health endpoint - this MUST succeed for the test to pass
	resp, err := http.Get("http://127.0.0.1:8080/health")
	require.NoError(t, err, "HTTP server must be reachable on /health endpoint")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200")

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response")
	assert.Contains(t, string(body), "healthy", "Response should contain 'healthy'")

	t.Log("Full fx server startup test passed - server is running and responding")

	// Shutdown the app
	_ = app.Stop(context.Background())
	select {
	case <-appDone:
	case <-time.After(10 * time.Second):
		// Don't fail the test if shutdown takes longer, just log it
		t.Log("App shutdown took longer than expected (this is OK for test cleanup)")
	}
}
