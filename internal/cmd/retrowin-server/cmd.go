// Package retrowinserver implements the retrowin-server command
package retrowinserver

import (
	"github.com/spf13/cobra"
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

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return ApplyMigrations(cfgFile)
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file path")

	return cmd
}
