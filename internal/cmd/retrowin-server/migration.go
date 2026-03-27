package retrowinserver

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql/schema"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/config"
)

// ApplyMigrations applies database migrations
func ApplyMigrations(cfg *config.Config) error {
	if cfg.App.Env == "development" {
		fmt.Printf("DSN: %s\n", cfg.DSN())
	}

	entClient, err := ent.Open(cfg.Database.Driver, cfg.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() {
		_ = entClient.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := entClient.Schema.Create(ctx, schema.WithGlobalUniqueID(true)); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	fmt.Println("Migrations applied successfully")
	return nil
}
