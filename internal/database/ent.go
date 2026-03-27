package database

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/config"
)

// Module provides the fx module for database.
var Module = fx.Module("database",
	fx.Provide(NewEntClient),
)

// NewEntClient creates a new Ent client.
func NewEntClient(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) (*ent.Client, error) {
	// Open database connection
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Create Ent driver
	drv := entsql.OpenDB(dialect.Postgres, db)

	// Create Ent client
	client := ent.NewClient(ent.Driver(drv))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Test connection
			if err := db.PingContext(ctx); err != nil {
				return fmt.Errorf("failed to ping database: %w", err)
			}
			logger.Info("connected to database",
				zap.String("host", cfg.Database.Host),
				zap.String("database", cfg.Database.Name),
			)

			// Auto migrate in development
			if cfg.App.Env == "development" {
				if err := client.Schema.Create(ctx); err != nil {
					logger.Warn("failed to auto-migrate schema", zap.Error(err))
				}
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database connection")
			if err := client.Close(); err != nil {
				return fmt.Errorf("failed to close ent client: %w", err)
			}
			return db.Close()
		},
	})

	return client, nil
}
