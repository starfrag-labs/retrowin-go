package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

// ProviderConfig holds OIDC provider configuration.
type ProviderConfig struct {
	ID           string
	Name         string
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Provider represents an OIDC provider.
type Provider struct {
	config ProviderConfig
}

// NewProvider creates a new OIDC provider.
func NewProvider(config ProviderConfig) *Provider {
	return &Provider{config: config}
}

// ID returns the provider ID.
func (p *Provider) ID() string {
	return p.config.ID
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return p.config.Name
}

// Issuer returns the issuer URL.
func (p *Provider) Issuer() string {
	return p.config.Issuer
}

// ClientID returns the client ID.
func (p *Provider) ClientID() string {
	return p.config.ClientID
}

// ClientSecret returns the client secret.
func (p *Provider) ClientSecret() string {
	return p.config.ClientSecret
}

// RedirectURI returns the redirect URI.
func (p *Provider) RedirectURI() string {
	return p.config.RedirectURI
}

// UserInfo contains user information from the OIDC provider.
type UserInfo struct {
	Subject       string
	Email         string
	EmailVerified bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
}

// LoginRequest represents an OIDC login request.
type LoginRequest struct {
	ProviderID   string `json:"providerId"`
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

// CallbackRequest represents an OIDC callback request.
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

// PKCE helpers

// GenerateCodeVerifier generates a PKCE code verifier.
func GenerateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GenerateCodeChallenge generates a PKCE code challenge from verifier.
func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
