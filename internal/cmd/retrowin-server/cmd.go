// Package retrowinserver implements the retrowin-server command
package retrowinserver

import (
	"github.com/spf13/cobra"

	"github.com/starfrag-lab/retrowin-go/internal/config"
)

// NewCmd creates a new cobra command for retrowin-server
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retrowin-server",
		Short: "RetroWin API server",
	}

	// Add subcommands
	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newMigrateCmd())

	return cmd
}

// newServeCmd creates the serve subcommand
func newServeCmd() *cobra.Command {
	var cfgFile string
	var port int

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		Run: func(cmd *cobra.Command, args []string) {
			app := NewFXApp(cfgFile, port)
			app.Run()
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "server port")

	return cmd
}

// newMigrateCmd creates the migrate command with subcommands
func newMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
	}

	cmd.AddCommand(newMigrateApplyCmd())

	return cmd
}

// newMigrateApplyCmd creates the migrate apply subcommand
func newMigrateApplyCmd() *cobra.Command {
	var cfgFile string
	var mode string
	var baseline string
	var clean bool
	var allowDirty bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg *config.Config
			var err error

			if cfgFile != "" {
				cfg, err = config.LoadFromPath(cfgFile)
			} else {
				cfg, err = config.Load("config.yaml")
			}
			if err != nil {
				return err
			}

			// CLI flags override config
			opts := MigrateOptions{
				Mode:       mode,
				Baseline:   baseline,
				Clean:      clean,
				AllowDirty: allowDirty,
			}
			if opts.Mode == "" {
				opts.Mode = "auto"
			}

			return ApplyMigrations(cfg, opts)
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	cmd.Flags().StringVar(&mode, "mode", "", "Migration mode: auto or versioned (overrides config)")
	cmd.Flags().StringVar(&baseline, "baseline", "", "Baseline version for existing databases (versioned mode)")
	cmd.Flags().BoolVar(&clean, "clean", false, "Drop all tables before applying migrations")
	cmd.Flags().BoolVar(&allowDirty, "allow-dirty", false, "Allow applying versioned migrations to a non-clean database")

	return cmd
}
