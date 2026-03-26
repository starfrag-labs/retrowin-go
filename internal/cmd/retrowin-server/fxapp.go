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
	"github.com/starfrag-lab/retrowin-go/internal/handler"
	"github.com/starfrag-lab/retrowin-go/internal/handler/v1"
	"github.com/starfrag-lab/retrowin-go/internal/storage"
	s3storage "github.com/starfrag-lab/retrowin-go/internal/storage/s3"
	"github.com/starfrag-lab/retrowin-go/internal/upload"
	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// FXApp wraps the fx application for easier use
type FXApp struct {
	app     *fx.App
	cfgFile string
	port    int
}

// NewFXApp creates a new fx application wrapper
func NewFXApp(cfgFile string, port int) *FXApp {
	return &FXApp{
		cfgFile: cfgFile,
		port:    port,
	}
}

// Run starts the fx application
func (a *FXApp) Run() {
	a.app = fx.New(FxOptions(a.cfgFile, a.port)...)
	a.app.Run()
}

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

// ProvideEntClient provides the ent database client.
func ProvideEntClient(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) (*ent.Client, error) {
	return database.NewEntClient(lc, cfg, logger)
}

// ProvideValkeyClient provides the Valkey client.
func ProvideValkeyClient(cfg *config.Config) (valkey.Client, error) {
	if cfg.Cache.Provider != "redis" && cfg.Cache.Provider != "valkey" {
		return nil, nil
	}

	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.Cache.Valkey.Addr},
		SelectDB:    cfg.Cache.Valkey.DB,
		Password:    cfg.Cache.Valkey.Password,
	})
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

// ProvideLogger creates a new zap logger.
func ProvideLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

// ProvideSecurityHandler creates the security handler for ogen.
func ProvideSecurityHandler(sessionSvc auth.SessionService) *v1.SecurityHandler {
	return v1.NewSecurityHandler(sessionSvc)
}

// Repository constructors using Ent
func ProvideUserRepository(client *ent.Client) user.Repository {
	return user.NewEntRepository(client)
}

func ProvideServiceStatusRepository(client *ent.Client) user.ServiceStatusRepository {
	return user.NewEntServiceStatusRepository(client)
}

func ProvideFileRepository(client *ent.Client) file.Repository {
	return file.NewEntRepository(client)
}

func ProvideFileInfoRepository(client *ent.Client) file.FileInfoRepository {
	return file.NewEntFileInfoRepository(client)
}

func ProvideFilePathRepository(client *ent.Client) file.FilePathRepository {
	return file.NewEntFilePathRepository(client)
}

func ProvideFileRoleRepository(client *ent.Client) file.FileRoleRepository {
	return file.NewEntFileRoleRepository(client)
}

// Service constructors
func ProvideUserService(userRepo user.Repository, statusRepo user.ServiceStatusRepository) user.Service {
	return user.NewService(userRepo, statusRepo)
}

func ProvideFileService(
	fileRepo file.Repository,
	infoRepo file.FileInfoRepository,
	pathRepo file.FilePathRepository,
	roleRepo file.FileRoleRepository,
) file.Service {
	return file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)
}

func ProvideStorage(cfg *config.Config) (storage.Storage, error) {
	return s3storage.New(&cfg.Storage)
}

func ProvideUploadService(fileSvc file.Service, storage storage.Storage) upload.Service {
	return upload.NewService(fileSvc, storage)
}

func ProvideHandler(
	userSvc user.Service,
	fileSvc file.Service,
	uploadSvc upload.Service,
	cfg *config.Config,
) *v1.Handler {
	return v1.NewHandler(userSvc, fileSvc, uploadSvc, cfg)
}

// Session constructors
func ProvideSessionRepository(client valkey.Client, cfg *config.Config) auth.SessionRepository {
	return auth.NewValkeySessionRepository(client, cfg.Auth.Session.RedisKey)
}

func ProvideSessionService(repo auth.SessionRepository, cfg *config.Config) auth.SessionService {
	ttl := time.Duration(cfg.Auth.Session.TTL) * time.Second
	return auth.NewSessionService(repo, ttl)
}

// Keycloak constructor
func ProvideKeycloak(cfg *config.Config) *auth.Keycloak {
	issuerURL := cfg.Auth.Keycloak.BaseURL + "/realms/" + cfg.Auth.Keycloak.Realm
	redirectURI := "http://localhost:8080/auth/callback" // TODO: Make configurable

	return auth.NewKeycloak(auth.KeycloakConfig{
		Issuer:       issuerURL,
		ClientID:     cfg.Auth.Keycloak.ClientID,
		ClientSecret: cfg.Auth.Keycloak.ClientSecret,
		RedirectURI:  redirectURI,
	})
}

func ProvideOIDCUserService(userSvc user.Service) auth.UserService {
	return auth.NewUserAdapter(userSvc, "keycloak")
}

func ProvideOIDCService(
	keycloak *auth.Keycloak,
	sessionSvc auth.SessionService,
	userSvc auth.UserService,
	client valkey.Client,
	cfg *config.Config,
) (auth.Service, error) {
	stateTTL := time.Duration(cfg.Auth.Session.StateTTL) * time.Second
	return auth.NewService(keycloak, sessionSvc, userSvc, client, cfg.Auth.Session.RedisKey, stateTTL)
}

func ProvideAuthHandler(oidcSvc auth.Service, sessionSvc auth.SessionService, cfg *config.Config) *handler.AuthHandler {
	return handler.NewAuthHandler(&handler.AuthHandlerConfig{
		AuthService:    oidcSvc,
		SessionService: sessionSvc,
		Secure:         cfg.Auth.Session.Secure,
	})
}

// Server represents the HTTP server.
type Server struct {
	ogenServer  *apiv1.Server
	authHandler *handler.AuthHandler
	config      *config.Config
	logger      *zap.Logger
}

// NewServer creates a new Server.
func NewServer(
	apiHandler *v1.Handler,
	securityHandler *v1.SecurityHandler,
	authHandler *handler.AuthHandler,
	cfg *config.Config,
	logger *zap.Logger,
) (*Server, error) {
	ogenServer, err := apiv1.NewServer(apiHandler, securityHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to create ogen server: %w", err)
	}

	return &Server{
		ogenServer:  ogenServer,
		authHandler: authHandler,
		config:      cfg,
		logger:      logger,
	}, nil
}

// ProvideHTTPServer provides the HTTP server.
func ProvideHTTPServer(
	srv *Server,
	cfg *config.Config,
	lc fx.Lifecycle,
) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)

	// Create a mux to combine API and auth routes
	mux := http.NewServeMux()

	// Mount auth routes
	mux.HandleFunc("/auth/login", srv.authHandler.Login)
	mux.HandleFunc("/auth/callback", srv.authHandler.Callback)
	mux.HandleFunc("/auth/logout", srv.authHandler.Logout)
	mux.HandleFunc("/auth/me", srv.authHandler.GetMe)

	// Serve OpenAPI spec
	mux.HandleFunc("/v1/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.HTTP.OpenAPIPath)
	})

	// Serve Swagger UI
	mux.Handle("/v1/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/v1/openapi.yaml"),
	))

	// Mount API routes
	mux.Handle("/", srv.ogenServer)

	httpSrv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", addr)
			if err != nil {
				return fmt.Errorf("failed to listen on %s: %w", addr, err)
			}

			srv.logger.Info("starting HTTP server", zap.String("addr", addr))

			go func() {
				if err := httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
					srv.logger.Error("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			srv.logger.Info("stopping HTTP server")
			return httpSrv.Shutdown(ctx)
		},
	})

	return httpSrv
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

// FxOptions returns the fx options for the application.
func FxOptions(cfgFile string, port int) []fx.Option {
	return []fx.Option{
		// Supply CLI args
		fx.Supply(fx.Annotate(cfgFile, fx.ResultTags(`name:"cfgFile"`))),
		fx.Supply(fx.Annotate(port, fx.ResultTags(`name:"port"`))),

		// Core providers
		fx.Provide(
			fx.Annotate(ProvideConfig, fx.ParamTags(`name:"cfgFile"`, `name:"port"`)),
			ProvideLogger,
			ProvideEntClient,
			ProvideValkeyClient,
		),

		// Repositories
		fx.Provide(
			ProvideUserRepository,
			ProvideServiceStatusRepository,
			ProvideFileRepository,
			ProvideFileInfoRepository,
			ProvideFilePathRepository,
			ProvideFileRoleRepository,
		),

		// Services
		fx.Provide(
			ProvideUserService,
			ProvideFileService,
			ProvideStorage,
			ProvideUploadService,
			ProvideHandler,
		),

		// Auth
		fx.Provide(
			ProvideSessionRepository,
			ProvideSessionService,
			ProvideKeycloak,
			ProvideOIDCUserService,
			ProvideOIDCService,
			ProvideAuthHandler,
			ProvideSecurityHandler,
		),

		// HTTP Server
		fx.Provide(NewServer),
		fx.Provide(ProvideHTTPServer),

		fx.Invoke(RegisterShutdownHooks),
		fx.Invoke(WaitForShutdown),
		fx.StartTimeout(30 * time.Second),
		fx.StopTimeout(30 * time.Second),
	}
}
