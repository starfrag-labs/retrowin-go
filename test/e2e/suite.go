// Package e2e provides end-to-end tests using testcontainers
package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	valkeymodule "github.com/testcontainers/testcontainers-go/modules/valkey"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/valkey-io/valkey-go"
	"go.uber.org/fx"
	"gopkg.in/yaml.v3"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/ent/object"
	"github.com/starfrag-lab/retrowin-go/ent/system"
	retrowinserver "github.com/starfrag-lab/retrowin-go/internal/cmd/retrowin-server"
	"github.com/starfrag-lab/retrowin-go/internal/config"
	domainsession "github.com/starfrag-lab/retrowin-go/internal/session"
	"github.com/starfrag-lab/retrowin-go/internal/session/repository"
)

// Suite holds the e2e test environment
type Suite struct {
	t               *testing.T
	PgContainer     *postgres.PostgresContainer
	ValkeyContainer *valkeymodule.ValkeyContainer
	Config          *config.Config
	EntClient       *ent.Client
	DB              *sql.DB
	ValkeyAddr      string
	ValkeyClient    valkey.Client

	// HTTP client with cookie management
	httpClient   *http.Client
	cookieJar    []*http.Cookie
	baseURL      string
	serverConfig *serverConfig
}

type serverConfig struct {
	cfgFile string
	port    int
	app     *fx.App
}

// NewSuite creates a new e2e test suite
func NewSuite(t *testing.T) *Suite {
	return &Suite{t: t}
}

// Start starts the test containers
func (s *Suite) Start(ctx context.Context) error {
	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("retrowin_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start postgres: %w", err)
	}
	s.PgContainer = pgContainer

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	// Start Valkey container
	valkeyContainer, err := valkeymodule.Run(ctx, "valkey/valkey:8-alpine",
		testcontainers.WithWaitStrategy(
			wait.ForLog("Ready to accept connections").
				WithOccurrence(1).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to start valkey: %w", err)
	}
	s.ValkeyContainer = valkeyContainer

	valkeyHost, err := valkeyContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get valkey host: %w", err)
	}

	valkeyPort, err := valkeyContainer.MappedPort(ctx, "6379")
	if err != nil {
		return fmt.Errorf("failed to get valkey port: %w", err)
	}
	s.ValkeyAddr = fmt.Sprintf("%s:%s", valkeyHost, valkeyPort.Port())

	// Create Valkey client for test helpers
	valkeyClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{s.ValkeyAddr},
	})
	if err != nil {
		return fmt.Errorf("failed to create valkey client: %w", err)
	}
	s.ValkeyClient = valkeyClient

	// Get postgres connection info
	pgHost, err := pgContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get postgres host: %w", err)
	}
	pgPort, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return fmt.Errorf("failed to get postgres port: %w", err)
	}

	// Connect to database
	s.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create Ent client
	drv := entsql.OpenDB(dialect.Postgres, s.DB)
	s.EntClient = ent.NewClient(ent.Driver(drv))

	// Run migrations
	if err := s.EntClient.Schema.Create(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}
	testPort := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()

	// Create test config with dynamically found port
	s.Config = &config.Config{
		App: config.AppConfig{
			Name:    "retrowin-test",
			Version: "test",
			Env:     "test",
		},
		HTTP: config.HTTPConfig{
			Host: "127.0.0.1",
			Port: testPort, // Use dynamically found available port
		},
		Database: config.DatabaseConfig{
			Driver:   "postgres",
			Host:     pgHost,
			Port:     pgPort.Int(),
			Name:     "retrowin_test",
			User:     "test",
			Password: "test",
			SSLMode:  "disable",
		},
		Cache: config.CacheConfig{
			Provider: "valkey",
			Valkey: config.ValkeyConfig{
				Addr:     s.ValkeyAddr,
				DB:       0,
				PoolSize: 10,
			},
		},
		Storage: config.StorageConfig{
			Provider: "memory",
			Bucket:   "test-bucket",
		},
		Auth: config.AuthConfig{
			Keycloak: config.KeycloakConfig{
				BaseURL:     "http://localhost:9999", // Invalid to prevent actual OIDC calls
				Realm:       "test",
				ClientID:    "test-client",
				RedirectURI: "http://localhost:18080/auth/callback",
			},
			Session: config.SessionConfig{
				TTL:      3600,
				Secure:   false,
				StateTTL: 300,
				RedisKey: "retrowin-test",
			},
		},
	}

	// Initialize HTTP client
	s.httpClient = &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}
	s.cookieJar = make([]*http.Cookie, 0)
	s.baseURL = fmt.Sprintf("http://%s:%d", s.Config.HTTP.Host, s.Config.HTTP.Port)

	return nil
}

// StartServer starts the actual HTTP server
func (s *Suite) StartServer(ctx context.Context) error {
	if s.serverConfig != nil {
		return nil // Already started
	}

	// Create a temporary config file
	tmpDir := s.t.TempDir()
	cfgFile := tmpDir + "/config.yaml"

	// Write config to temp file as YAML
	cfgData, err := yaml.Marshal(s.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	err = os.WriteFile(cfgFile, cfgData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	s.t.Logf("Using config file: %s", cfgFile)
	s.t.Logf("Server will listen on: %s:%d", s.Config.HTTP.Host, s.Config.HTTP.Port)

	// Start the actual fx app
	app := retrowinserver.NewFXApp(cfgFile, s.Config.HTTP.Port)

	appDone := make(chan struct{})
	go func() {
		app.Run()
		close(appDone)
	}()

	// Wait for server to be ready
	s.t.Log("Waiting for server to start...")
	for i := 0; i < 30; i++ {
		resp, err := http.Get(s.baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				s.t.Log("Server is ready")
				break
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Verify server is running
	resp, err := http.Get(s.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server health check failed: %d", resp.StatusCode)
	}

	s.serverConfig = &serverConfig{
		cfgFile: cfgFile,
		port:    s.Config.HTTP.Port,
		app:     app,
	}

	return nil
}

// Stop stops the test environment
func (s *Suite) Stop(ctx context.Context) error {
	// Stop server first with a fresh context to ensure it has time to shut down
	if s.serverConfig != nil && s.serverConfig.app != nil {
		// Use a fresh context with timeout for shutdown
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.serverConfig.app.Stop(stopCtx)
	}

	if s.ValkeyClient != nil {
		s.ValkeyClient.Close()
	}
	if s.EntClient != nil{
		_ = s.EntClient.Close()
	}
	if s.DB != nil{
		_ = s.DB.Close()
	}

	if s.PgContainer != nil{
		if err := testcontainers.TerminateContainer(s.PgContainer); err != nil{
			s.t.Logf("Failed to terminate postgres: %v", err)
		}
	}

	if s.ValkeyContainer != nil{
		if err := testcontainers.TerminateContainer(s.ValkeyContainer); err != nil{
			s.t.Logf("Failed to terminate valkey: %v", err)
		}
	}

	return nil
}

// HTTP Methods

// Get performs a GET request
func (s *Suite) Get(path string) (*http.Response, error) {
	return s.Do("GET", path, nil)
}

// Post performs a POST request with JSON body
func (s *Suite) Post(path string, body interface{}) (*http.Response, error) {
	return s.Do("POST", path, body)
}

// Put performs a PUT request with JSON body
func (s *Suite) Put(path string, body interface{}) (*http.Response, error) {
	return s.Do("PUT", path, body)
}

// Patch performs a PATCH request with JSON body
func (s *Suite) Patch(path string, body interface{}) (*http.Response, error) {
	return s.Do("PATCH", path, body)
}

// Delete performs a DELETE request
func (s *Suite) Delete(path string) (*http.Response, error) {
	return s.Do("DELETE", path, nil)
}

// Do performs an HTTP request
func (s *Suite) Do(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	fullURL := s.baseURL + path
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add cookies from jar
	for _, cookie := range s.cookieJar {
		req.AddCookie(cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Save cookies from response
	s.cookieJar = append(s.cookieJar, resp.Cookies()...)

	return resp, nil
}

// ReadJSON reads JSON from response body
func (s *Suite) ReadJSON(resp *http.Response, v interface{}) error {
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w (body: %s)", err, string(data))
	}
	return nil
}

// ReadBody reads response body as string
func (s *Suite) ReadBody(resp *http.Response) string {
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

// ClearCookies clears the cookie jar
func (s *Suite) ClearCookies() {
	s.cookieJar = make([]*http.Cookie, 0)
}

// AddCookie adds a cookie to the jar
func (s *Suite) AddCookie(cookie *http.Cookie) {
	s.cookieJar = append(s.cookieJar, cookie)
}

// Test Data Helpers

// CreateTestUser creates a test user in the database
func (s *Suite) CreateTestUser(ctx context.Context, provider, providerID, username string) (*ent.User, error) {
	u, err := s.EntClient.User.Create().
		SetID(uuid.New().String()).
		SetProvider(provider).
		SetProviderID(providerID).
		SetUsername(username).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}
	return u, nil
}

// CreateTestSystem creates a test system in the database
func (s *Suite) CreateTestSystem(ctx context.Context, name string) (*ent.System, error) {
	sys, err := s.EntClient.System.Create().
		SetID(uuid.New().String()).
		SetName(name).
		SetStatus(system.StatusActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create test system: %w", err)
	}
	return sys, nil
}

// CreateTestSession creates a test session for a user (bypasses auth)
func (s *Suite) CreateTestSession(ctx context.Context, userID string) (string, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

	// Create session in Valkey - use same prefix as server
	sessionRepo := repository.NewValkeySessionRepository(s.ValkeyClient, "retrowin:session:")
	session := domainsession.NewSession(
		domainsession.SessionID(sessionID),
		userID,
		expiresAt,
		now,
	)

	if err := sessionRepo.Save(ctx, session); err != nil {
		return "", fmt.Errorf("failed to create test session: %w", err)
	}

	return sessionID, nil
}

// CreateTestSystemUser creates a system user (UID/GID)
func (s *Suite) CreateTestSystemUser(ctx context.Context, systemID, userID, username string, uid, gid int) (*ent.UserSystem, error) {
	su, err := s.EntClient.UserSystem.Create().
		SetSystemID(systemID).
		SetUserID(userID).
		SetUsername(username).
		SetUID(uid).
		SetGid(gid).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create test system user: %w", err)
	}
	return su, nil
}

// CreateTestSystemGroup creates a system group
func (s *Suite) CreateTestSystemGroup(ctx context.Context, systemID, name string, gid int) (*ent.SystemGroup, error) {
	sg, err := s.EntClient.SystemGroup.Create().
		SetSystemID(systemID).
		SetName(name).
		SetGid(gid).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create test system group: %w", err)
	}
	return sg, nil
}

// CreateTestInode creates a test inode
func (s *Suite) CreateTestInode(ctx context.Context, systemID string, mode, uid, gid int, content []byte) (*ent.Inode, error) {
	in, err := s.EntClient.Inode.Create().
		SetSystemID(systemID).
		SetMode(mode).
		SetUID(uid).
		SetGid(gid).
		SetSize(int64(len(content))).
		SetLinkCount(1).
		SetContent(content).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create test inode: %w", err)
	}
	return in, nil
}

// CreateTestObject creates a test object
func (s *Suite) CreateTestObject(ctx context.Context, systemID, bucket, storageKey string) (*ent.Object, error) {
	obj, err := s.EntClient.Object.Create().
		SetID(uuid.New().String()).
		SetSystemID(systemID).
		SetProvider(object.ProviderS3).
		SetBucket(bucket).
		SetStorageKey(storageKey).
		SetStatus(object.StatusActive).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create test object: %w", err)
	}
	return obj, nil
}

// LoginAs logs in as a user by creating a session and setting cookies
func (s *Suite) LoginAs(ctx context.Context, userID string) error {
	sessionID, err := s.CreateTestSession(ctx, userID)
	if err != nil {
		return err
	}

	// Set session cookie
	sessionCookie := &http.Cookie{
		Name:  "retrowin_session",
		Value: sessionID,
	}
	s.cookieJar = append(s.cookieJar, sessionCookie)

	return nil
}

// Logout clears session cookies
func (s *Suite) Logout() {
	s.ClearCookies()
}

// SetupAuthenticatedUser creates a user and logs in as them
func (s *Suite) SetupAuthenticatedUser(ctx context.Context, username string) (*ent.User, error) {
	provider := "test"
	providerID := fmt.Sprintf("test-%s", username)

	u, err := s.CreateTestUser(ctx, provider, providerID, username)
	if err != nil {
		return nil, err
	}

	if err := s.LoginAs(ctx, u.ID); err != nil {
		return nil, err
	}

	return u, nil
}

// SetupSystemWithUser creates a system and adds the user to it
func (s *Suite) SetupSystemWithUser(ctx context.Context, systemName string, userID, username string) (*ent.System, *ent.UserSystem, error) {
	sys, err := s.CreateTestSystem(ctx, systemName)
	if err != nil {
		return nil, nil, err
	}

	// Create root group
	_, err = s.CreateTestSystemGroup(ctx, sys.ID, "root", 0)
	if err != nil {
		return nil, nil, err
	}

	// Create user's private group
	_, err = s.CreateTestSystemGroup(ctx, sys.ID, username, 1000)
	if err != nil {
		return nil, nil, err
	}

	// Create system user
	su, err := s.CreateTestSystemUser(ctx, sys.ID, userID, username, 1000, 1000)
	if err != nil {
		return nil, nil, err
	}

	return sys, su, nil
}

// SetupFullEnvironment creates a complete test environment with user, system, and session
func (s *Suite) SetupFullEnvironment(ctx context.Context, username string) (*ent.User, *ent.System, *ent.UserSystem, error) {
	// Create user
	u, err := s.SetupAuthenticatedUser(ctx, username)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup authenticated user: %w", err)
	}

	// Create system with user
	sys, su, err := s.SetupSystemWithUser(ctx, fmt.Sprintf("%s-system", username), u.ID, username)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup system: %w", err)
	}

	return u, sys, su, nil
}

// BuildURL builds a URL with path parameters
func (s *Suite) BuildURL(path string, params ...interface{}) string {
	return s.baseURL + fmt.Sprintf(path, params...)
}

// BuildURLWithQuery builds a URL with query parameters
func (s *Suite) BuildURLWithQuery(path string, query url.Values) string {
	if len(query) == 0 {
		return s.baseURL + path
	}
	return s.baseURL + path + "?" + query.Encode()
}

// CleanupDatabase cleans all test data from the database
func (s *Suite) CleanupDatabase(ctx context.Context) error {
	// Delete in correct order to handle foreign keys
	_, _ = s.EntClient.Inode.Delete().Exec(ctx)
	_, _ = s.EntClient.Object.Delete().Exec(ctx)
	_, _ = s.EntClient.UserGroup.Delete().Exec(ctx)
	_, _ = s.EntClient.SystemGroup.Delete().Exec(ctx)
	_, _ = s.EntClient.UserSystem.Delete().Exec(ctx)
	_, _ = s.EntClient.System.Delete().Exec(ctx)
	_, _ = s.EntClient.User.Delete().Exec(ctx)
	return nil
}

// CleanupValkey clears all test data from Valkey
func (s *Suite) CleanupValkey(ctx context.Context) error {
	// Delete all keys with our prefix
	return s.ValkeyClient.Do(ctx, s.ValkeyClient.B().Del().Key("retrowin-test:*").Build()).Error()
}

// AssertStatusCode asserts the HTTP status code
func (s *Suite) AssertStatusCode(expected, actual int, msgAndArgs ...interface{}) {
	require.Equal(s.t, expected, actual, msgAndArgs...)
}

// AssertJSONContentType asserts the Content-Type is application/json
func (s *Suite) AssertJSONContentType(resp *http.Response) {
	require.Equal(s.t, "application/json", resp.Header.Get("Content-Type"), "Content-Type should be application/json")
}
