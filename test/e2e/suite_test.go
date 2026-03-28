package e2e

import (
	"context"
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
