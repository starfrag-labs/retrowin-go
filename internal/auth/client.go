package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Client wraps OIDC/OAuth2 operations.
type Client struct {
	provider     *Provider
	oidcProvider *oidc.Provider
	oauth2Config *oauth2.Config
}

// NewClient creates a new OIDC client.
func NewClient(ctx context.Context, provider *Provider) (*Client, error) {
	oidcProvider, err := oidc.NewProvider(ctx, provider.Issuer())
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     provider.ClientID(),
		ClientSecret: provider.ClientSecret(),
		RedirectURL:  provider.RedirectURI(),
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &Client{
		provider:     provider,
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
	return c.oidcProvider.Verifier(&oidc.Config{ClientID: c.provider.ClientID()}).Verify(ctx, rawIDToken)
}
