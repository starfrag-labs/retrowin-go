// Package server implements the retrowin-server command
package retrowinserver

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"go.uber.org/fx"
	"go.uber.org/zap"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/valkey-io/valkey-go"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/application/storage"
	"github.com/starfrag-lab/retrowin-go/internal/auth"
	"github.com/starfrag-lab/retrowin-go/internal/config"
	corefs "github.com/starfrag-lab/retrowin-go/internal/core/fs"
	"github.com/starfrag-lab/retrowin-go/internal/core/inode"
	inoderepo "github.com/starfrag-lab/retrowin-go/internal/core/inode/repository"
	"github.com/starfrag-lab/retrowin-go/internal/core/object"
	objectrepo "github.com/starfrag-lab/retrowin-go/internal/core/object/repository"
	s3storage "github.com/starfrag-lab/retrowin-go/internal/core/object/s3"
	coreuser "github.com/starfrag-lab/retrowin-go/internal/core/user"
	"github.com/starfrag-lab/retrowin-go/internal/handler"
	"github.com/starfrag-lab/retrowin-go/internal/service/sysinit"
	"github.com/starfrag-lab/retrowin-go/internal/session"
	sessionRepo "github.com/starfrag-lab/retrowin-go/internal/session/repository"
	"github.com/starfrag-lab/retrowin-go/internal/system"
	systemrepo "github.com/starfrag-lab/retrowin-go/internal/system/repository"
	"github.com/starfrag-lab/retrowin-go/internal/user"
	userrepo "github.com/starfrag-lab/retrowin-go/internal/user/repository"
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

// NewEntClient creates a new Ent client.
func NewEntClient(lc fx.Lifecycle, cfg *config.Config, logger *zap.Logger) (*ent.Client, error) {
	// Open database connection
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Create Ent driver
	drv := entsql.OpenDB(dialect.Postgres, db)

	// Create Ent client
	client := ent.NewClient(ent.Driver(drv))

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Test connection
			if err := db.PingContext(ctx); err != nil {
				return fmt.Errorf("failed to ping database: %w", err)
			}
			logger.Info("connected to database",
				zap.String("host", cfg.Database.Host),
				zap.String("database", cfg.Database.Name),
			)

			// Auto migrate in development
			if cfg.App.Env == "development" {
				if err := client.Schema.Create(ctx); err != nil {
					logger.Warn("failed to auto-migrate schema", zap.Error(err))
				}
			}

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database connection")
			if err := client.Close(); err != nil {
				return fmt.Errorf("failed to close ent client: %w", err)
			}
			return db.Close()
		},
	})

	return client, nil
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

// ProvideSessionTTL provides session TTL from config.
func ProvideSessionTTL(cfg *config.Config) time.Duration {
	return time.Duration(cfg.Auth.Session.TTL) * time.Second
}

// ProvideAuthUserService provides the auth user service.
func ProvideAuthUserService(userSvc user.UserService) auth.UserService {
	return auth.NewUserService(userSvc)
}

// ProvideStorage provides the storage backend based on config.
func ProvideStorage(cfg *config.Config) (object.Storage, error) {
	return s3storage.New(&cfg.Storage)
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
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	// API routes (no prefix)
	mux.Handle("/", ogenServer)

	// Serve OpenAPI spec and Swagger UI
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.HTTP.OpenAPIPath)
	})
	mux.HandleFunc("/swagger", httpSwagger.Handler(
		httpSwagger.URL("/openapi.json"),
	))

	return mux
}

// logoutMiddleware wraps the handler to clear session cookie on logout.
func logoutMiddleware(next http.Handler, secure bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For logout requests, set the cookie-clearing header before the handler runs
		if r.Method == http.MethodPost && r.URL.Path == "/auth/logout" {
			// Clear the session cookie before calling the handler
			http.SetCookie(w, &http.Cookie{
				Name:     "retrowin_session",
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				Secure:   secure,
				MaxAge:   -1,
			})
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// ProvideHTTPHandler provides the HTTP handler.
func ProvideHTTPHandler(mux *http.ServeMux, cfg *config.Config) http.Handler {
	return logoutMiddleware(mux, cfg.Auth.Session.Secure)
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
// Taking srv *http.Server as parameter ensures the HTTP server provider is constructed.
func RegisterShutdownHooks(lc fx.Lifecycle, entClient *ent.Client, valkeyCli valkey.Client, srv *http.Server) {
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
	sessionSvc session.SessionService,
) (*apiv1.Server, error) {
	securityHandler := handler.NewSecurityHandler(sessionSvc)
	return apiv1.NewServer(h, securityHandler, apiv1.WithErrorHandler(h.ErrorHandler))
}

// FxOptions returns the fx options for the application.
func FxOptions(cfgFile string, port int) []fx.Option {
	return []fx.Option{
		// Supply CLI args
		fx.Supply(fx.Annotate(cfgFile, fx.ResultTags(`name:"cfgFile"`))),
		fx.Supply(fx.Annotate(port, fx.ResultTags(`name:"port"`))),

		// All providers - single fx.Provide call like serengeti
		fx.Provide(
			fx.Annotate(ProvideConfig, fx.ParamTags(`name:"cfgFile"`, `name:"port"`)),
			ProvideLogger,
			NewEntClient,
			ProvideValkeyClient,
			// Repositories
			userrepo.NewRepository,
			inoderepo.NewRepository,
			objectrepo.NewRepository,
			systemrepo.NewSystemUserRepository,
			systemrepo.NewSystemGroupRepository,
			systemrepo.NewRepository,
			NewValkeySessionRepository,
			ProvideSessionTTL,
			// Auth services
			session.NewSessionService,
			ProvideAuthUserService,
			// OIDC service
			ProvideKeycloak,
			ProvideOIDCService,
			// Domain services
			user.NewService,
			inode.NewService,
			object.NewService,
			coreuser.NewService,      // core/user for UID resolution
			coreuser.NewGroupService, // core/user for group management
			system.NewService,        // system management
			sysinit.NewService,       // system initialization
			// Application services
			corefs.NewService,
			storage.NewService,
			// Storage
			ProvideStorage,
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

// NewValkeySessionRepository provides the Valkey session repository.
func NewValkeySessionRepository(client valkey.Client) session.SessionRepository {
	if client == nil {
		return nil
	}
	return sessionRepo.NewValkeySessionRepository(client, "retrowin:session:")
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
	sessionSvc session.SessionService,
	userSvc auth.UserService,
	client valkey.Client,
	cfg *config.Config,
) (auth.AuthService, error) {
	stateTTL := time.Duration(cfg.Auth.Session.StateTTL) * time.Second
	return auth.NewService(keycloak, sessionSvc, userSvc, client, cfg.Auth.Session.RedisKey, stateTTL)
}
