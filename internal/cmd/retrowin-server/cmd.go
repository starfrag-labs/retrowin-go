// Package retrowinserver implements the retrowin-server command
package retrowinserver

import (
	"github.com/spf13/cobra"

	"github.com/starfrag-lab/retrowin-go/internal/cmd/gc"
	"github.com/starfrag-lab/retrowin-go/internal/cmd/migrate"
	"github.com/starfrag-lab/retrowin-go/internal/cmd/serve"
)

// NewCmd creates a new cobra command for retrowin-server
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retrowin-server",
		Short: "RetroWin API server",
	}

	// Add subcommands from separate packages
	cmd.AddCommand(serve.NewCmd())
	cmd.AddCommand(migrate.NewCmd())
	cmd.AddCommand(gc.NewCmd())

	return cmd
}
