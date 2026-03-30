package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
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

// Client wraps OIDC/OAuth2 operations for Keycloak.
type Client struct {
	keycloak     *Keycloak
	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
}

// NewClient creates a new OIDC client for Keycloak.
func NewClient(ctx context.Context, keycloak *Keycloak) (*Client, error) {
	oidcProvider, err := oidc.NewProvider(ctx, keycloak.Issuer())
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:    keycloak.ClientID(),
		RedirectURL: keycloak.RedirectURI(),
		Endpoint:    oidcProvider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}
	if keycloak.ClientSecret() != "" {
		oauth2Config.ClientSecret = keycloak.ClientSecret()
	}

	return &Client{
		keycloak:     keycloak,
		oidcProvider: oidcProvider,
		oauth2Config: oauth2Config,
	}, nil
}

// AuthURL generates an authorization URL with PKCE.
func (c *Client) AuthURL(state, codeChallenge string) string {
	return c.oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)
}

// Exchange exchanges an authorization code for tokens.
func (c *Client) Exchange(ctx context.Context, code, codeVerifier string) (*oauth2.Token, error) {
	return c.oauth2Config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
}

// GetUserInfo retrieves user info from the provider.
func (c *Client) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	info, err := c.oidcProvider.UserInfo(ctx, c.oauth2Config.TokenSource(ctx, token))
	if err != nil {
		return nil, err
	}

	var claims struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}

	if err := info.Claims(&claims); err != nil {
		return nil, err
	}

	return &UserInfo{
		Subject:       claims.Sub,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		GivenName:     claims.GivenName,
		FamilyName:    claims.FamilyName,
		Picture:       claims.Picture,
	}, nil
}

// VerifyToken verifies an ID token.
func (c *Client) VerifyToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return c.oidcProvider.Verifier(&oidc.Config{ClientID: c.keycloak.ClientID()}).Verify(ctx, rawIDToken)
}

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
