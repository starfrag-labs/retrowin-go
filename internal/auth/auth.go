package auth

import (
	"time"
)

// UserInfo contains user information from Keycloak.
type UserInfo struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
}

// LoginRequest represents a Keycloak login request.
type LoginRequest struct {
	RedirectURI  string `json:"redirectUri"`
	State        string `json:"state"`
	CodeVerifier string `json:"codeVerifier"`
}

// LoginResponse contains the authorization URL for login.
type LoginResponse struct {
	AuthorizationURL string    `json:"authorizationUrl"`
	State            string    `json:"state"`
	CodeVerifier     string    `json:"codeVerifier"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

// CallbackRequest represents a Keycloak callback request.
type CallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// CallbackResponse contains the session after successful login.
type CallbackResponse struct {
	SessionID string    `json:"sessionId"`
	UserID    int64     `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}
