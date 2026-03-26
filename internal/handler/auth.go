package handler

import (
	"encoding/json"
	"net/http"

	"github.com/starfrag-lab/retrowin-go/internal/auth"
	"github.com/starfrag-lab/retrowin-go/internal/auth/session"
	"github.com/starfrag-lab/retrowin-go/internal/errors"
	"github.com/starfrag-lab/retrowin-go/internal/middleware"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authSvc    auth.Service
	sessionSvc session.Service
	secure     bool
}

// AuthHandlerConfig holds configuration for the auth handler.
type AuthHandlerConfig struct {
	AuthService    auth.Service
	SessionService session.Service
	Secure         bool
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(cfg *AuthHandlerConfig) *AuthHandler {
	return &AuthHandler{
		authSvc:    cfg.AuthService,
		sessionSvc: cfg.SessionService,
		secure:     cfg.Secure,
	}
}

// Login initiates the OIDC login flow.
// GET /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	resp, err := h.authSvc.InitiateLogin(r.Context())
	if err != nil {
		writeAuthError(w, errors.Internal(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Callback handles the OIDC callback.
// GET /auth/callback
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		writeAuthError(w, errors.BadRequest("missing code or state"))
		return
	}

	req := &auth.CallbackRequest{
		Code:  code,
		State: state,
	}

	resp, err := h.authSvc.HandleCallback(r.Context(), req)
	if err != nil {
		writeAuthError(w, errors.Unauthorized(err.Error()))
		return
	}

	// Set session cookie
	middleware.SetSessionCookie(w, resp.SessionID, h.secure)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Logout handles user logout.
// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Delete session from storage
	_ = h.sessionSvc.Delete(r.Context(), session.ID(cookie.Value))

	// Clear session cookie
	middleware.ClearSessionCookie(w, h.secure)

	w.WriteHeader(http.StatusNoContent)
}

// GetMe returns the current user info.
// GET /auth/me
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(middleware.SessionCookieName)
	if err != nil {
		writeAuthError(w, errors.Unauthorized("not authenticated"))
		return
	}

	sess, err := h.sessionSvc.Validate(r.Context(), session.ID(cookie.Value))
	if err != nil {
		writeAuthError(w, errors.Unauthorized("invalid or expired session"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"userId":    sess.UserID(),
		"expiresAt": sess.ExpiresAt(),
	})
}

func writeAuthError(w http.ResponseWriter, err *errors.Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.StatusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"code":    err.Code,
		"message": err.Message,
	})
}
