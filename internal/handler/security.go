package handler

import (
	"context"

	"github.com/ogen-go/ogen/ogenerrors"

	api "github.com/starfrag-lab/retrowin-go/pkg/api"

	"github.com/starfrag-lab/retrowin-go/internal/session"
	"github.com/starfrag-lab/retrowin-go/internal/utils"
)

// SecurityHandler implements the ogen SecurityHandler interface.
type SecurityHandler struct {
	sessionSvc session.SessionService
}

// NewSecurityHandler creates a new SecurityHandler.
func NewSecurityHandler(sessionSvc session.SessionService) *SecurityHandler {
	return &SecurityHandler{
		sessionSvc: sessionSvc,
	}
}

// HandleSessionAuth handles sessionAuth security.
func (h *SecurityHandler) HandleSessionAuth(ctx context.Context, operationName api.OperationName, t api.SessionAuth) (context.Context, error) {
	sess, err := h.sessionSvc.Validate(ctx, session.SessionID(t.APIKey))
	if err != nil {
		return ctx, ogenerrors.ErrSkipServerSecurity
	}

	// Add user ID to context
	ctx = utils.ContextWithUserID(ctx, sess.UserID())

	return ctx, nil
}
