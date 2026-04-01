package handler

import (
	"context"
	"time"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/application/storage"
	"github.com/starfrag-lab/retrowin-go/internal/auth"
	corefs "github.com/starfrag-lab/retrowin-go/internal/core/fs"
	coreuser "github.com/starfrag-lab/retrowin-go/internal/core/user"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/service/sysinit"
	"github.com/starfrag-lab/retrowin-go/internal/system"
	"github.com/starfrag-lab/retrowin-go/internal/utils"
	extuser "github.com/starfrag-lab/retrowin-go/internal/user"
)

// Handler implements the ogen API handler interface.
type Handler struct {
	// Auth service
	authSvc auth.AuthService

	// External user service (for /user endpoints)
	extUserSvc extuser.UserService

	// System user/group services (for /systems/{id}/users and /systems/{id}/groups endpoints)
	sysUserSvc  coreuser.UserService
	sysGroupSvc coreuser.GroupService

	// System service
	systemSvc system.SystemService

	// Filesystem and storage services
	fsSvc      corefs.FsService
	storageSvc storage.StorageService

	// Init service
	initSvc sysinit.InitService
}

// NewHandler creates a new Handler.
func NewHandler(
	authSvc auth.AuthService,
	extUserSvc extuser.UserService,
	sysUserSvc coreuser.UserService,
	sysGroupSvc coreuser.GroupService,
	systemSvc system.SystemService,
	fsSvc corefs.FsService,
	storageSvc storage.StorageService,
	initSvc sysinit.InitService,
) *Handler {
	return &Handler{
		authSvc:     authSvc,
		extUserSvc:  extUserSvc,
		sysUserSvc:  sysUserSvc,
		sysGroupSvc: sysGroupSvc,
		systemSvc:   systemSvc,
		fsSvc:       fsSvc,
		storageSvc:  storageSvc,
		initSvc:     initSvc,
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

// domainError converts domain errors to HTTP errors.
func (h *Handler) domainError(err error) error {
	return err // ErrorHandler/NewError in errors.go handles status code mapping
}

// checkSystemAccess verifies that the authenticated user has access to the given system.
func (h *Handler) checkSystemAccess(ctx context.Context, systemID string) error {
	userID, ok := utils.GetUserID(ctx)
	if !ok {
		return errors.Unauthorized("user not authenticated")
	}

	_, err := h.sysUserSvc.GetByUserAndSystem(ctx, userID, systemID)
	if err != nil {
		return errors.Forbidden("access denied to system")
	}
	return nil
}
