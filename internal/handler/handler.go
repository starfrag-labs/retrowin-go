package handler

import (
	"context"
	"time"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/application/storage"
	"github.com/starfrag-lab/retrowin-go/internal/auth"
	"github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/user"
	initsvc "github.com/starfrag-lab/retrowin-go/internal/service/init"
	"github.com/starfrag-lab/retrowin-go/internal/system"
)

// Handler implements the ogen API handler interface.
type Handler struct {
	authSvc   auth.AuthService
	userSvc   user.UserService
	systemSvc system.SystemService
	fsSvc     fs.FsService
	storageSvc storage.StorageService
	initSvc   initsvc.InitService
}

// NewHandler creates a new Handler.
func NewHandler(
	authSvc auth.AuthService,
	userSvc user.UserService,
	systemSvc system.SystemService,
	fsSvc fs.FsService,
	storageSvc storage.StorageService,
	initSvc initsvc.InitService,
) *Handler {
	return &Handler{
		authSvc:   authSvc,
		userSvc:   userSvc,
		systemSvc: systemSvc,
		fsSvc:     fsSvc,
		storageSvc: storageSvc,
		initSvc:   initSvc,
	}
}

// Ensure Handler implements the ogen Handler interface.
var _ apiv1.Handler = (*Handler)(nil)

// Helper functions

func toTimestamp(t time.Time) apiv1.Timestamp {
	return apiv1.Timestamp(t)
}

func toOptTimestamp(t time.Time) apiv1.OptTimestamp {
	return apiv1.NewOptTimestamp(toTimestamp(t))
}
