package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// KeycloakConfig holds Keycloak configuration.
type KeycloakConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// Keycloak represents the Keycloak OIDC provider.
type Keycloak struct {
	config KeycloakConfig
}

// NewKeycloak creates a new Keycloak provider.
func NewKeycloak(config KeycloakConfig) *Keycloak {
	return &Keycloak{config: config}
}

// Issuer returns the issuer URL.
func (k *Keycloak) Issuer() string {
	return k.config.Issuer
}

// ClientID returns the client ID.
func (k *Keycloak) ClientID() string {
	return k.config.ClientID
}

// ClientSecret returns the client secret.
func (k *Keycloak) ClientSecret() string {
	return k.config.ClientSecret
}

// RedirectURI returns the redirect URI.
func (k *Keycloak) RedirectURI() string {
	return k.config.RedirectURI
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
