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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

// Suite represents an independent e2e test environment
type Suite struct {
	t               *testing.T
	PgContainer     *postgres.PostgresContainer
	ValkeyContainer *valkeymodule.ValkeyContainer
	MinioContainer  testcontainers.Container
	Config          *config.Config
	EntClient       *ent.Client
	DB              *sql.DB
	ValkeyAddr      string
	MinioAddr       string
	ValkeyClient    valkey.Client
	baseURL         string
	httpClient      *http.Client
	cookieJar       []*http.Cookie
	app             *fx.App
	cfgFile         string
}

// NewSuite creates a new independent test suite
func NewSuite(t *testing.T) *Suite {
	return &Suite{
		t: t,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		cookieJar: make([]*http.Cookie, 0),
	}
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

	// Start MinIO container
	minioContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "minio/minio:latest",
			ExposedPorts: []string{"9000/tcp"},
			Env: map[string]string{
				"MINIO_ROOT_USER":     "minioadmin",
				"MINIO_ROOT_PASSWORD": "minioadmin",
			},
			Cmd: []string{"server", "/data"},
			WaitingFor: wait.ForExposedPort().
				WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return fmt.Errorf("failed to start minio: %w", err)
	}
	s.MinioContainer = minioContainer

	minioHost, err := minioContainer.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get minio host: %w", err)
	}
	minioPort, err := minioContainer.MappedPort(ctx, "9000")
	if err != nil {
		return fmt.Errorf("failed to get minio port: %w", err)
	}
	s.MinioAddr = fmt.Sprintf("%s:%s", minioHost, minioPort.Port())

	// Create test bucket in MinIO
	if err := s.createMinioBucket(ctx); err != nil {
		return fmt.Errorf("failed to create minio bucket: %w", err)
	}

	// Create test config with dynamically found port
	s.Config = &config.Config{
		App: config.AppConfig{
			Name:    "retrowin-test",
			Version: "test",
			Env:     "test",
		},
		HTTP: config.HTTPConfig{
			Host:        "127.0.0.1",
			Port:        testPort,
			OpenAPIPath: "../../api/openapi.yaml",
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
			Provider:  "s3",
			Region:    "us-east-1",
			Endpoint:  "http://" + s.MinioAddr,
			AccessKey: "minioadmin",
			SecretKey: "minioadmin",
			Bucket:    "test-bucket",
		},
		Auth: config.AuthConfig{
			Keycloak: config.KeycloakConfig{
				BaseURL:     "http://localhost:9999",
				Realm:       "test",
				ClientID:    "test-client",
				RedirectURI: "http://localhost:18080/auth/callback",
			},
			Session: config.SessionConfig{
				TTL:        3600,
				Secure:     false,
				StateTTL:   300,
				RedisKey:   "retrowin-test",
				CookieName: "session_id",
			},
		},
	}

	s.baseURL = fmt.Sprintf("http://%s:%d", s.Config.HTTP.Host, s.Config.HTTP.Port)

	return nil
}

// StartServer starts the HTTP server
func (s *Suite) StartServer(ctx context.Context) error {
	// Create a temporary config file
	tmpDir, err := os.MkdirTemp("", "retrowin-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	s.cfgFile = tmpDir + "/config.yaml"

	// Write config to temp file as YAML
	cfgData, err := yaml.Marshal(s.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	err = os.WriteFile(s.cfgFile, cfgData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Start the actual fx app
	s.app = retrowinserver.NewFXApp(s.cfgFile, s.Config.HTTP.Port)

	go func() {
		s.app.Run()
	}()

	// Wait for server to be ready
	for i := 0; i < 30; i++ {
		resp, err := http.Get(s.baseURL + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
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

	return nil
}

// Stop stops everything
func (s *Suite) Stop(ctx context.Context) error {
	if s.app != nil {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.app.Stop(stopCtx)
	}

	if s.ValkeyClient != nil {
		s.ValkeyClient.Close()
	}
	if s.EntClient != nil {
		_ = s.EntClient.Close()
	}
	if s.DB != nil {
		_ = s.DB.Close()
	}

	if s.PgContainer != nil {
		_ = testcontainers.TerminateContainer(s.PgContainer)
	}

	if s.ValkeyContainer != nil {
		_ = testcontainers.TerminateContainer(s.ValkeyContainer)
	}

	if s.MinioContainer != nil {
		_ = testcontainers.TerminateContainer(s.MinioContainer)
	}

	return nil
}

// BaseURL returns the base URL
func (s *Suite) BaseURL() string {
	return s.baseURL
}

// GetEntClient returns the Ent client
func (s *Suite) GetEntClient() *ent.Client {
	return s.EntClient
}

// GetDB returns the database connection
func (s *Suite) GetDB() *sql.DB {
	return s.DB
}

// GetConfig returns the test config
func (s *Suite) GetConfig() *config.Config {
	return s.Config
}

// GetPgContainer returns the postgres container
func (s *Suite) GetPgContainer() *postgres.PostgresContainer {
	return s.PgContainer
}

// HTTP Methods

func (s *Suite) Get(path string) (*http.Response, error) {
	return s.Do("GET", path, nil)
}

func (s *Suite) Post(path string, body any) (*http.Response, error) {
	return s.Do("POST", path, body)
}

func (s *Suite) Put(path string, body any) (*http.Response, error) {
	return s.Do("PUT", path, body)
}

func (s *Suite) Patch(path string, body any) (*http.Response, error) {
	return s.Do("PATCH", path, body)
}

func (s *Suite) Delete(path string) (*http.Response, error) {
	return s.Do("DELETE", path, nil)
}

func (s *Suite) Do(method, path string, body any) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = strings.NewReader(string(jsonData))
	}

	fullURL := s.BaseURL() + path
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	for _, cookie := range s.cookieJar {
		req.AddCookie(cookie)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	s.cookieJar = append(s.cookieJar, resp.Cookies()...)

	return resp, nil
}

func (s *Suite) ReadJSON(resp *http.Response, v any) error {
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

func (s *Suite) ReadBody(resp *http.Response) string {
	defer func() { _ = resp.Body.Close() }()
	data, _ := io.ReadAll(resp.Body)
	return string(data)
}

func (s *Suite) ClearCookies() {
	s.cookieJar = make([]*http.Cookie, 0)
}

func (s *Suite) AddCookie(cookie *http.Cookie) {
	s.cookieJar = append(s.cookieJar, cookie)
}

// Test Data Helpers

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

func (s *Suite) CreateTestSession(ctx context.Context, userID string) (string, error) {
	sessionID := uuid.New().String()
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)

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

func (s *Suite) LoginAs(ctx context.Context, userID string) error {
	sessionID, err := s.CreateTestSession(ctx, userID)
	if err != nil {
		return err
	}
	sessionCookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
	}
	s.cookieJar = append(s.cookieJar, sessionCookie)
	return nil
}

func (s *Suite) Logout() {
	s.ClearCookies()
}

func (s *Suite) SetupAuthenticatedUser(ctx context.Context, username string) (*ent.User, error) {
	provider := "keycloak"
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

func (s *Suite) SetupSystemWithUser(ctx context.Context, systemName string, userID, username string) (*ent.System, *ent.UserSystem, error) {
	sys, err := s.CreateTestSystem(ctx, systemName)
	if err != nil {
		return nil, nil, err
	}

	_, err = s.CreateTestSystemGroup(ctx, sys.ID, "root", 0)
	if err != nil {
		return nil, nil, err
	}

	_, err = s.CreateTestSystemGroup(ctx, sys.ID, username, 1000)
	if err != nil {
		return nil, nil, err
	}

	su, err := s.CreateTestSystemUser(ctx, sys.ID, userID, username, 1000, 1000)
	if err != nil {
		return nil, nil, err
	}

	return sys, su, nil
}

func (s *Suite) SetupFullEnvironment(ctx context.Context, username string) (*ent.User, *ent.System, *ent.UserSystem, error) {
	u, err := s.SetupAuthenticatedUser(ctx, username)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup authenticated user: %w", err)
	}

	sys, su, err := s.SetupSystemWithUser(ctx, fmt.Sprintf("%s-system", username), u.ID, username)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to setup system: %w", err)
	}

	return u, sys, su, nil
}

// SetupFullEnvironmentAPI creates a system via the API (which initializes filesystem)
// and returns the created system. This is the preferred method for e2e tests that need
// a properly initialized filesystem.
func (s *Suite) SetupFullEnvironmentAPI(ctx context.Context, username string) (*ent.User, map[string]any, error) {
	// Setup authenticated user
	u, err := s.SetupAuthenticatedUser(ctx, username)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup authenticated user: %w", err)
	}

	// Create system via API (this initializes filesystem with root directory, /home, etc.)
	req := map[string]any{
		"name":        fmt.Sprintf("%s-system", username),
		"description": "Test system for e2e tests",
	}

	resp, err := s.Post("/systems", req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create system: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		body := s.ReadBody(resp)
		return nil, nil, fmt.Errorf("failed to create system: status=%d body=%s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := s.ReadJSON(resp, &result); err != nil {
		// Re-read since ReadJSON closes body
		return nil, nil, fmt.Errorf("failed to parse system response: %w", err)
	}

	// Note: resp.Body is already closed by ReadJSON, so we don't defer close

	return u, result, nil
}

func (s *Suite) BuildURL(path string, params ...any) string {
	return s.BaseURL() + fmt.Sprintf(path, params...)
}

func (s *Suite) BuildURLWithQuery(path string, query url.Values) string {
	if len(query) == 0 {
		return s.BaseURL() + path
	}
	return s.BaseURL() + path + "?" + query.Encode()
}

// createMinioBucket creates the test bucket in MinIO.
func (s *Suite) createMinioBucket(ctx context.Context) error {
	client, err := minio.New(s.MinioAddr, &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	if err := client.MakeBucket(ctx, "test-bucket", minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// UploadToPresignedURL uploads data to a presigned URL (for testing MinIO uploads).
func (s *Suite) UploadToPresignedURL(t *testing.T, presignedURL string, data []byte) {
	t.Logf("Uploading to presigned URL: %s", presignedURL)

	req, err := http.NewRequest("PUT", presignedURL, strings.NewReader(string(data)))
	require.NoError(t, err, "Failed to create upload request")
	req.Header.Set("Content-Type", "text/plain")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to upload to presigned URL")
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Upload to presigned URL failed with status %d: %s\nURL: %s", resp.StatusCode, string(body), presignedURL)
	}
}
