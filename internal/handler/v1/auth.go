package v1

import (
	"context"
	"net/url"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/core/auth"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
)

// InitiateLogin implements GET /auth/login.
func (h *Handler) InitiateLogin(ctx context.Context) (*apiv1.LoginResponse, error) {
	resp, err := h.authSvc.InitiateLogin(ctx)
	if err != nil {
		return nil, err
	}

	authURL, err := url.Parse(resp.AuthorizationURL)
	if err != nil {
		return nil, err
	}

	return &apiv1.LoginResponse{
		AuthorizationUrl: *authURL,
		State:            resp.State,
	}, nil
}

// HandleCallback implements POST /auth/callback.
func (h *Handler) HandleCallback(ctx context.Context, req *apiv1.CallbackRequest) (apiv1.HandleCallbackRes, error) {
	callbackReq := &auth.CallbackRequest{
		Code:  req.Code,
		State: req.State,
	}

	resp, err := h.authSvc.HandleCallback(ctx, callbackReq)
	if err != nil {
		return &apiv1.HandleCallbackUnauthorized{}, nil
	}

	return &apiv1.CallbackResponse{
		SessionId: resp.SessionID,
		UserId:    resp.UserID,
		ExpiresAt: apiv1.OptTimestamp{Value: apiv1.Timestamp(resp.ExpiresAt), Set: true},
	}, nil
}

// Logout implements POST /auth/logout.
func (h *Handler) Logout(ctx context.Context) (apiv1.LogoutRes, error) {
	sessionID := middleware.GetSessionID(ctx)
	if sessionID == "" {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "unauthorized",
				Message: "no session found",
			},
		}, nil
	}

	err := h.authSvc.Logout(ctx, sessionID)
	if err != nil {
		return &apiv1.Error{
			Error: apiv1.ErrorError{
				Type:    "unauthorized",
				Message: "logout failed",
			},
		}, nil
	}

	return &apiv1.LogoutNoContent{}, nil
}
