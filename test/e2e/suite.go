// Package e2e provides end-to-end tests using testcontainers
package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/valkey"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/config"
)

// Suite holds the e2e test environment
type Suite struct {
	t               *testing.T
	PgContainer     *postgres.PostgresContainer
	ValkeyContainer testcontainers.Container
	Config          *config.Config
	EntClient       *ent.Client
	DB              *sql.DB
	ValkeyAddr      string
}

// NewSuite creates a new e2e test suite
func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

// Start starts the test containers
func (s *Suite) Start(ctx context.Context) error {
	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("retrowin_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres: %w", err)
	}
	s.PgContainer = pgContainer

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	// Start Valkey container
	valkeyContainer, err := valkey.Run(ctx, "valkey/valkey:8-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start valkey: %w", err)
	}
	s.ValkeyContainer = valkeyContainer

	valkeyHost, err := valkeyContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get valkey host: %w", err)
	}

	valkeyPort, err := valkeyContainer.MappedPort(ctx, "6379")
	if err != nil {
		return fmt.Errorf("failed to get valkey port: %w", err)
	}
	s.ValkeyAddr = fmt.Sprintf("%s:%s", valkeyHost, valkeyPort.Port())

	// Connect to database
	s.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create Ent client
	drv := entsql.OpenDB(dialect.Postgres, s.DB)
	s.EntClient = ent.NewClient(ent.Driver(drv))

	// Run migrations
	if err := s.EntClient.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create test config
	s.Config = &config.Config{
		App: config.AppConfig{
			Name:    "retrowin-test",
			Version: "test",
			Env:     "test",
		},
		HTTP: config.HTTPConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			Driver:   "postgres",
			Host:     "",
			Port:     5432,
			Name:     "retrowin_test",
			User:     "test",
			Password: "test",
			SSLMode:  "disable",
		},
		Cache: config.CacheConfig{
			Provider: "valkey",
			Valkey: config.ValkeyConfig{
				Addr:     s.ValkeyAddr,
				DB:       0,
				PoolSize: 10,
			},
		},
		Storage: config.StorageConfig{
			Provider: "s3",
			Region:   "us-east-1",
			Bucket:   "test-bucket",
		},
		Auth: config.AuthConfig{
			Keycloak: config.KeycloakConfig{
				BaseURL:     "http://localhost:8080",
				Realm:       "test",
				ClientID:    "test-client",
				RedirectURI: "http://localhost:8080/auth/callback",
			},
			Session: config.SessionConfig{
				TTL:      3600,
				Secure:   false,
				StateTTL: 300,
				RedisKey: "retrowin-test",
			},
		},
	}

	return nil
}

// Stop stops the test environment
func (s *Suite) Stop(ctx context.Context) error {
	if s.EntClient != nil {
		_ = s.EntClient.Close()
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}

	if s.PgContainer != nil {
		if err := testcontainers.TerminateContainer(s.PgContainer); err != nil {
			s.t.Logf("Failed to terminate postgres: %v", err)
		}
	}

	if s.ValkeyContainer != nil {
		if err := testcontainers.TerminateContainer(s.ValkeyContainer); err != nil {
			s.t.Logf("Failed to terminate valkey: %v", err)
		}
	}

	return nil
}
