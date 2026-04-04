package handler

import (
	"context"
	"net/url"

	api "github.com/starfrag-lab/retrowin-go/pkg/api"

	"github.com/starfrag-lab/retrowin-go/internal/auth"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
)

// InitiateLogin implements GET /auth/login.
func (h *Handler) InitiateLogin(ctx context.Context) (api.InitiateLoginRes, error) {
	resp, err := h.authSvc.InitiateLogin(ctx)
	if err != nil {
		return nil, err
	}

	authURL, err := url.Parse(resp.AuthorizationURL)
	if err != nil {
		return nil, err
	}

	return &api.LoginResponse{
		AuthorizationUrl: *authURL,
		State:            resp.State,
	}, nil
}

// HandleCallback implements GET /auth/callback.
func (h *Handler) HandleCallback(ctx context.Context, params api.HandleCallbackParams) (api.HandleCallbackRes, error) {
	callbackReq := &auth.CallbackRequest{
		Code:  params.Code,
		State: params.State,
	}

	resp, err := h.authSvc.HandleCallback(ctx, callbackReq)
	if err != nil {
		if errors.IsUnauthorized(err) || errors.IsNotFound(err) {
			return &api.HandleCallbackUnauthorized{}, nil
		}
		return &api.HandleCallbackBadRequest{}, nil
	}

	return &api.CallbackResponse{
		SessionId: resp.SessionID,
		UserId:    resp.UserID,
		ExpiresAt: api.OptTimestamp{Value: api.Timestamp(resp.ExpiresAt), Set: true},
	}, nil
}

// Logout implements POST /auth/logout.
func (h *Handler) Logout(ctx context.Context) error {
	sessionID := middleware.GetSessionID(ctx)
	if sessionID == "" {
		// No session - return success (idempotent logout)
		return nil
	}

	err := h.authSvc.Logout(ctx, sessionID)
	if err != nil {
		// Log error but still return success (idempotent logout)
		// The session might have already been deleted or expired
		return nil
	}

	return nil
}
