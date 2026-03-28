// Package server implements the retrowin-server command
package retrowinserver

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/valkey-io/valkey-go"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/auth"
	"github.com/starfrag-lab/retrowin-go/internal/config"
	"github.com/starfrag-lab/retrowin-go/internal/database"
	"github.com/starfrag-lab/retrowin-go/internal/file"
	handler "github.com/starfrag-lab/retrowin-go/internal/handler/v1"
	"github.com/starfrag-lab/retrowin-go/internal/storage"
	s3storage "github.com/starfrag-lab/retrowin-go/internal/storage/s3"
	"github.com/starfrag-lab/retrowin-go/internal/upload"
	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// ProvideConfig provides the config from file.
func ProvideConfig(cfgFile string, port int) (*config.Config, error) {
	var cfg *config.Config
	var err error

	if cfgFile != "" {
		cfg, err = config.LoadFromPath(cfgFile)
	} else {
		cfg, err = config.Load("config.yaml")
	}
	if err != nil {
		return nil, err
	}

	// Override port if specified
	if port != 8080 {
		cfg.HTTP.Port = port
	}

	return cfg, nil
}

// ProvideLogger creates a new zap logger.
func ProvideLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

// ProvideValkeyClient provides the Valkey client.
func ProvideValkeyClient(cfg *config.Config) (valkey.Client, error) {
	if cfg.Cache.Provider != "redis" && cfg.Cache.Provider != "valkey" {
		return nil, nil
	}

	client, err := newValkeyClient(&cfg.Cache.Valkey)
	if err != nil {
		return nil, fmt.Errorf("failed to create valkey client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to cache: %w", err)
	}
	fmt.Printf("Connected to %s\n", cfg.Cache.Provider)

	return client, nil
}

// ProvideRepositories provides all repository implementations.
func ProvideRepositories(
	entClient *ent.Client,
	valkeyCli valkey.Client,
) struct {
	UserRepo    user.Repository
	StatusRepo  user.ServiceStatusRepository
	FileRepo    file.Repository
	InfoRepo    file.FileInfoRepository
	PathRepo    file.FilePathRepository
	RoleRepo    file.FileRoleRepository
	SessionRepo auth.SessionRepository
} {
	userRepo := user.NewEntRepository(entClient)
	statusRepo := user.NewEntServiceStatusRepository(entClient)
	fileRepo := file.NewEntRepository(entClient)
	infoRepo := file.NewEntFileInfoRepository(entClient)
	pathRepo := file.NewEntFilePathRepository(entClient)
	roleRepo := file.NewEntFileRoleRepository(entClient)

	var sessionRepo auth.SessionRepository
	if valkeyCli != nil {
		sessionRepo = auth.NewValkeySessionRepository(valkeyCli, "retrowin:session:")
	} else {
		// TODO: Implement memory session repository
		sessionRepo = nil
	}

	return struct {
		UserRepo    user.Repository
		StatusRepo  user.ServiceStatusRepository
		FileRepo    file.Repository
		InfoRepo    file.FileInfoRepository
		PathRepo    file.FilePathRepository
		RoleRepo    file.FileRoleRepository
		SessionRepo auth.SessionRepository
	}{
		UserRepo:    userRepo,
		StatusRepo:  statusRepo,
		FileRepo:    fileRepo,
		InfoRepo:    infoRepo,
		PathRepo:    pathRepo,
		RoleRepo:    roleRepo,
		SessionRepo: sessionRepo,
	}
}

// ProvideSessionTTL provides session TTL from config.
func ProvideSessionTTL(cfg *config.Config) time.Duration {
	return time.Duration(cfg.Auth.Session.TTL) * time.Second
}

// ProvideHTTPMux provides the HTTP mux with all routes.
func ProvideHTTPMux(
	ogenServer *apiv1.Server,
	cfg *config.Config,
) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint (direct access, no /v1 prefix)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes with /v1 prefix
	mux.Handle("/v1/", http.StripPrefix("/v1", ogenServer))

	// Serve OpenAPI spec and Swagger UI
	mux.HandleFunc("/v1/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.HTTP.OpenAPIPath)
	})
	mux.HandleFunc("/v1/swagger", httpSwagger.Handler(
		httpSwagger.URL("/v1/openapi.json"),
	))

	return mux
}

// ProvideHTTPHandler provides the HTTP handler.
func ProvideHTTPHandler(mux *http.ServeMux) http.Handler {
	return mux
}

// ProvideHTTPServer provides the HTTP server.
func ProvideHTTPServer(
	handler http.Handler,
	cfg *config.Config,
	lc fx.Lifecycle,
) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("failed to listen on %s: %w", addr, err)
			}

			fmt.Printf("Starting server on %s\n", addr)

			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					fmt.Printf("Server error: %v\n", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			fmt.Println("Shutting down server...")
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

// RegisterShutdownHooks registers shutdown signal handlers.
func RegisterShutdownHooks(lc fx.Lifecycle, entClient *ent.Client, valkeyCli valkey.Client) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			if err := entClient.Close(); err != nil {
				return fmt.Errorf("database close error: %w", err)
			}
			if valkeyCli != nil {
				valkeyCli.Close()
			}
			fmt.Println("Server stopped")
			return nil
		},
	})
}

// WaitForShutdown waits for shutdown signals.
func WaitForShutdown(lc fx.Lifecycle) {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				sig := <-shutdownCh
				fmt.Printf("\nReceived signal %v, shutting down...\n", sig)
			}()
			return nil
		},
	})
}

// ProvideOgenServer provides the ogen server.
func ProvideOgenServer(
	h *handler.Handler,
	sessionSvc auth.SessionService,
) (*apiv1.Server, error) {
	securityHandler := handler.NewSecurityHandler(sessionSvc)
	return apiv1.NewServer(h, securityHandler)
}

// FxOptions returns the fx options for the application.
func FxOptions(cfgFile string, port int) []fx.Option {
	return []fx.Option{
		// Supply CLI args
		fx.Supply(fx.Annotate(cfgFile, fx.ResultTags(`name:"cfgFile"`))),
		fx.Supply(fx.Annotate(port, fx.ResultTags(`name:"port"`))),

		// Core providers
		database.Module, // fx.Module must be passed directly, not inside fx.Provide
		fx.Provide(
			fx.Annotate(ProvideConfig, fx.ParamTags(`name:"cfgFile"`, `name:"port"`)),
			ProvideLogger,
			ProvideValkeyClient,
			ProvideRepositories,
			ProvideSessionTTL,
		),

		// Services - fx resolves dependencies automatically from constructors
		fx.Provide(
			// Auth services
			auth.NewSessionService,
			auth.NewUserAdapter,

			// OIDC service
			ProvideKeycloak,
			ProvideOIDCService,

			// Domain services
			user.NewService,
			file.NewService,
			ProvideStorage,
			upload.NewService,

			// HTTP layer
			handler.NewHandler,
			ProvideOgenServer,
			ProvideHTTPMux,
			ProvideHTTPHandler,
			ProvideHTTPServer,
		),

		fx.Invoke(RegisterShutdownHooks),
		fx.Invoke(WaitForShutdown),
		fx.StartTimeout(30 * time.Second),
		fx.StopTimeout(30 * time.Second),
	}
}

// NewFXApp creates a new fx application.
func NewFXApp(cfgFile string, port int) *fx.App {
	return fx.New(FxOptions(cfgFile, port)...)
}

// newValkeyClient creates a Valkey client based on ValkeyConfig.
func newValkeyClient(cfg *config.ValkeyConfig) (valkey.Client, error) {
	opts := valkey.ClientOption{
		InitAddress: []string{cfg.Addr},
	}
	if cfg.Password != "" {
		opts.Password = cfg.Password
	}
	if cfg.DB > 0 {
		opts.SelectDB = cfg.DB
	}
	if cfg.PoolSize > 0 {
		opts.BlockingPoolSize = cfg.PoolSize
	}

	return valkey.NewClient(opts)
}

// ProvideKeycloak provides the Keycloak OIDC client.
func ProvideKeycloak(cfg *config.Config) *auth.Keycloak {
	issuerURL := cfg.Auth.Keycloak.BaseURL + "/realms/" + cfg.Auth.Keycloak.Realm
	return auth.NewKeycloak(auth.KeycloakConfig{
		Issuer:       issuerURL,
		ClientID:     cfg.Auth.Keycloak.ClientID,
		ClientSecret: cfg.Auth.Keycloak.ClientSecret,
		RedirectURI:  cfg.Auth.Keycloak.RedirectURI,
	})
}

// ProvideOIDCService provides the OIDC service.
func ProvideOIDCService(
	keycloak *auth.Keycloak,
	sessionSvc auth.SessionService,
	userSvc auth.UserService,
	client valkey.Client,
	repos struct {
		SessionRepo auth.SessionRepository
	},
	cfg *config.Config,
) (auth.Service, error) {
	stateTTL := time.Duration(cfg.Auth.Session.StateTTL) * time.Second
	return auth.NewService(keycloak, sessionSvc, userSvc, client, cfg.Auth.Session.RedisKey, stateTTL)
}

// ProvideStorage provides the S3 storage.
func ProvideStorage(cfg *config.Config) (storage.Storage, error) {
	return s3storage.New(&cfg.Storage)
}
