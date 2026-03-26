package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

// UserService defines the interface for user operations.
type UserService interface {
	// FindOrCreate finds an existing user by OIDC subject or creates a new one.
	FindOrCreate(ctx context.Context, subject, email, name, picture string) (int64, error)
}

// Service defines the authentication service interface.
type Service interface {
	// GetKeycloak returns the Keycloak provider.
	GetKeycloak() *Keycloak

	// GetClient returns the OIDC client.
	GetClient() *Client

	// InitiateLogin starts the OIDC login flow.
	InitiateLogin(ctx context.Context) (*LoginResponse, error)

	// HandleCallback handles the OIDC callback.
	HandleCallback(ctx context.Context, req *CallbackRequest) (*CallbackResponse, error)

	// ValidateState validates the OAuth state parameter.
	ValidateState(ctx context.Context, state string) (*LoginRequest, error)
}

type authService struct {
	keycloak     *Keycloak
	client       *Client
	sessionSvc   SessionService
	userSvc      UserService
	valkey       *redis.Client
	valkeyPrefix string
	stateTTL     time.Duration
}

// NewService creates a new authentication service.
func NewService(
	keycloak *Keycloak,
	sessionSvc SessionService,
	userSvc UserService,
	valkey *redis.Client,
	valkeyPrefix string,
	stateTTL time.Duration,
) (Service, error) {
	client, err := NewClient(context.Background(), keycloak)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC client: %w", err)
	}

	return &authService{
		keycloak:     keycloak,
		client:       client,
		sessionSvc:   sessionSvc,
		userSvc:      userSvc,
		valkey:       valkey,
		valkeyPrefix: valkeyPrefix,
		stateTTL:     stateTTL,
	}, nil
}

// GetKeycloak returns the Keycloak provider.
func (s *authService) GetKeycloak() *Keycloak {
	return s.keycloak
}

// GetClient returns the OIDC client.
func (s *authService) GetClient() *Client {
	return s.client
}

// InitiateLogin starts the OIDC login flow.
func (s *authService) InitiateLogin(ctx context.Context) (*LoginResponse, error) {
	// Generate PKCE verifier and challenge
	codeVerifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := GenerateCodeChallenge(codeVerifier)

	// Generate state
	state, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// Store login request in Valkey
	loginReq := &LoginRequest{
		RedirectURI:  s.keycloak.RedirectURI(),
		State:        state,
		CodeVerifier: codeVerifier,
	}

	if err := s.storeLoginRequest(ctx, loginReq); err != nil {
		return nil, fmt.Errorf("failed to store login request: %w", err)
	}

	// Generate authorization URL
	authURL := s.client.AuthURL(state, codeChallenge)

	return &LoginResponse{
		AuthorizationURL: authURL,
		State:            state,
		CodeVerifier:     codeVerifier,
		ExpiresAt:        time.Now().Add(s.stateTTL),
	}, nil
}

// HandleCallback handles the OIDC callback.
func (s *authService) HandleCallback(ctx context.Context, req *CallbackRequest) (*CallbackResponse, error) {
	// Validate state
	loginReq, err := s.ValidateState(ctx, req.State)
	if err != nil {
		return nil, err
	}

	// Exchange code for tokens
	token, err := s.client.Exchange(ctx, req.Code, loginReq.CodeVerifier)
	if err != nil {
		return nil, errors.Unauthorized(fmt.Sprintf("failed to exchange code: %v", err))
	}

	// Get user info
	userInfo, err := s.client.GetUserInfo(ctx, token)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to get user info: %v", err))
	}

	// Find or create user
	userID, err := s.userSvc.FindOrCreate(
		ctx,
		userInfo.Subject,
		userInfo.Email,
		userInfo.Name,
		userInfo.Picture,
	)
	if err != nil {
		return nil, err
	}

	// Create session
	sess, err := s.sessionSvc.Create(ctx, userID)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to create session: %v", err))
	}

	// Delete used login request
	s.deleteLoginRequest(ctx, req.State)

	return &CallbackResponse{
		SessionID: string(sess.ID()),
		UserID:    userID,
		ExpiresAt: sess.ExpiresAt(),
	}, nil
}

// ValidateState validates the OAuth state parameter.
func (s *authService) ValidateState(ctx context.Context, state string) (*LoginRequest, error) {
	loginReq, err := s.getLoginRequest(ctx, state)
	if err != nil {
		return nil, errors.Unauthorized("invalid or expired state")
	}
	return loginReq, nil
}

// storeLoginRequest stores the login request in Valkey.
func (s *authService) storeLoginRequest(ctx context.Context, req *LoginRequest) error {
	key := s.getStateKey(req.State)
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return s.valkey.Set(ctx, key, data, s.stateTTL).Err()
}

// getLoginRequest retrieves the login request from Valkey.
func (s *authService) getLoginRequest(ctx context.Context, state string) (*LoginRequest, error) {
	key := s.getStateKey(state)
	data, err := s.valkey.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, errors.NotFound("login request not found")
	}
	if err != nil {
		return nil, err
	}

	var req LoginRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

// deleteLoginRequest deletes the login request from Valkey.
func (s *authService) deleteLoginRequest(ctx context.Context, state string) {
	key := s.getStateKey(state)
	s.valkey.Del(ctx, key)
}

func (s *authService) getStateKey(state string) string {
	return fmt.Sprintf("%s:auth:state:%s", s.valkeyPrefix, state)
}
