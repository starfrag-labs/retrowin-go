// Package serve implements the serve command
package serve

import (
	"github.com/spf13/cobra"
)

// NewCmd creates a new cobra command for the serve command
func NewCmd() *cobra.Command {
	var cfgFile string
	var port int
	var openAPIPath string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the API server",
		Run: func(cmd *cobra.Command, args []string) {
			app := NewFXApp(cfgFile, port, openAPIPath)
			app.Run()
		},
	}

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file path")
	cmd.Flags().IntVarP(&port, "port", "p", 8080, "server port")
	cmd.Flags().StringVar(&openAPIPath, "openapi", "", "OpenAPI spec file path")

	return cmd
}
