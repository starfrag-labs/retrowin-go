package v1

import (
	"context"

	"github.com/ogen-go/ogen/ogenerrors"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/middleware"
	"github.com/starfrag-lab/retrowin-go/internal/session"
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
func (h *SecurityHandler) HandleSessionAuth(ctx context.Context, operationName apiv1.OperationName, t apiv1.SessionAuth) (context.Context, error) {
	sess, err := h.sessionSvc.Validate(ctx, session.SessionID(t.APIKey))
	if err != nil {
		return ctx, ogenerrors.ErrSkipServerSecurity
	}

	// Add user ID to context
	ctx = context.WithValue(ctx, middleware.UserIDKey, sess.UserID())

	return ctx, nil
}
