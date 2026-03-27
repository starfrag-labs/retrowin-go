package retrowinserver

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/config"
)

// ApplyMigrations applies database migrations
func ApplyMigrations(cfgFile string) error {
	// Load config
	var cfg *config.Config
	var err error

	if cfgFile != "" {
		cfg, err = config.LoadFromPath(cfgFile)
	} else {
		cfg, err = config.Load("config.yaml")
	}
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.App.Env == "development" {
		fmt.Printf("DSN: %s\n", cfg.DSN())
	}

	// Create ent client
	entClient, err := ent.Open(cfg.Database.Driver, cfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		_ = entClient.Close()
	}()

	// Run migrations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	opts := []schema.MigrateOption{
		schema.WithGlobalUniqueID(true),
	}

	if err := entClient.Schema.Create(ctx, opts...); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}
