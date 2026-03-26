package main

import (
	"go.uber.org/fx"

	"github.com/starfrag-lab/retrowin-go/ent"
	retrowinserver "github.com/starfrag-lab/retrowin-go/internal/cmd/retrowin-server"
	"github.com/starfrag-lab/retrowin-go/internal/database"
)

func main() {
	app := fx.New(
		database.Module,
		fx.Provide(func(client *ent.Client) *ent.Client { return client }),
		retrowinserver.Module,
	)

	app.Run()
}
