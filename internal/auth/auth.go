package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"

	"github.com/starfrag-lab/retrowin-go/internal/errors"
)

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
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// AuthService defines the authentication service interface.
type AuthService interface {
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

	// Logout deletes the session.
	Logout(ctx context.Context, sessionID string) error
}

type authService struct {
	keycloak     *Keycloak
	client       *Client
	sessionSvc   SessionService
	userSvc      UserService
	valkey       valkey.Client
	valkeyPrefix string
	stateTTL     time.Duration
}

// NewService creates a new authentication service.
func NewService(
	keycloak *Keycloak,
	sessionSvc SessionService,
	userSvc UserService,
	valkey valkey.Client,
	valkeyPrefix string,
	stateTTL time.Duration,
) (AuthService, error) {
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
	codeVerifier, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := GenerateCodeChallenge(codeVerifier)

	state, err := GenerateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	loginReq := &LoginRequest{
		RedirectURI:  s.keycloak.RedirectURI(),
		State:        state,
		CodeVerifier: codeVerifier,
	}

	if err := s.storeLoginRequest(ctx, loginReq); err != nil {
		return nil, fmt.Errorf("failed to store login request: %w", err)
	}

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
	loginReq, err := s.ValidateState(ctx, req.State)
	if err != nil {
		return nil, err
	}

	token, err := s.client.Exchange(ctx, req.Code, loginReq.CodeVerifier)
	if err != nil {
		return nil, errors.Unauthorized(fmt.Sprintf("failed to exchange code: %v", err))
	}

	userInfo, err := s.client.GetUserInfo(ctx, token)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to get user info: %v", err))
	}

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

	sess, err := s.sessionSvc.Create(ctx, userID)
	if err != nil {
		return nil, errors.Internal(fmt.Sprintf("failed to create session: %v", err))
	}

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

// Logout deletes the session.
func (s *authService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionSvc.Delete(ctx, SessionID(sessionID))
}

func (s *authService) storeLoginRequest(ctx context.Context, req *LoginRequest) error {
	key := s.getStateKey(req.State)
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return s.valkey.Do(ctx, s.valkey.B().Set().
		Key(key).
		Value(string(data)).
		ExSeconds(int64(s.stateTTL.Seconds())).
		Build()).Error()
}

func (s *authService) getLoginRequest(ctx context.Context, state string) (*LoginRequest, error) {
	key := s.getStateKey(state)
	result := s.valkey.Do(ctx, s.valkey.B().Get().Key(key).Build())
	if result.Error() != nil {
		return nil, errors.NotFound("login request not found")
	}

	data, err := result.ToString()
	if err != nil {
		return nil, errors.NotFound("login request not found")
	}

	var req LoginRequest
	if err := json.Unmarshal([]byte(data), &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func (s *authService) deleteLoginRequest(ctx context.Context, state string) {
	key := s.getStateKey(state)
	_ = s.valkey.Do(ctx, s.valkey.B().Del().Key(key).Build()).Error()
}

func (s *authService) getStateKey(state string) string {
	return fmt.Sprintf("%s:auth:state:%s", s.valkeyPrefix, state)
}
