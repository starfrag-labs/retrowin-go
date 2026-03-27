package main

import (
	"os"

	retrowinserver "github.com/starfrag-lab/retrowin-go/internal/cmd/retrowin-server"
)

var version = "dev"

func main() {
	cmd := retrowinserver.NewCmd()
	cmd.Version = version
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
