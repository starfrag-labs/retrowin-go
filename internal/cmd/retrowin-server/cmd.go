package retrowinserver

import (
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/valkey-io/valkey-go"

	"github.com/starfrag-lab/retrowin-go/ent"
	"github.com/starfrag-lab/retrowin-go/internal/auth"
	"github.com/starfrag-lab/retrowin-go/internal/config"
	"github.com/starfrag-lab/retrowin-go/internal/file"
	"github.com/starfrag-lab/retrowin-go/internal/handler"
	"github.com/starfrag-lab/retrowin-go/internal/handler/v1"
	"github.com/starfrag-lab/retrowin-go/internal/storage"
	s3storage "github.com/starfrag-lab/retrowin-go/internal/storage/s3"
	"github.com/starfrag-lab/retrowin-go/internal/upload"
	"github.com/starfrag-lab/retrowin-go/internal/user"
)

// Module provides the fx module for the retrowin server.
var Module = fx.Module("retrowin",
	fx.Provide(
		NewLogger,
		NewConfig,
		NewValkeyClient,
		NewSecurityHandler,
		NewSessionRepository,
		NewSessionService,
		NewKeycloak,
		NewOIDCUserService,
		NewOIDCService,
		NewAuthHandler,
		NewUserService,
		NewFileService,
		NewStorage,
		NewUploadService,
		NewHandler,
		NewServer,
		// Ent Repositories
		NewUserRepository,
		NewServiceStatusRepository,
		NewFileRepository,
		NewFileInfoRepository,
		NewFilePathRepository,
		NewFileRoleRepository,
	),
	fx.Invoke(func(*Server) {}),
)

// NewLogger creates a new zap logger.
func NewLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

// NewConfig loads the configuration.
func NewConfig() (*config.Config, error) {
	// TODO: Make config path configurable
	return config.Load("config.yaml")
}

// NewSecurityHandler creates the security handler for ogen.
func NewSecurityHandler(sessionSvc auth.SessionService) *v1.SecurityHandler {
	return v1.NewSecurityHandler(sessionSvc)
}

// Repository constructors using Ent

func NewUserRepository(client *ent.Client) user.Repository {
	return user.NewEntRepository(client)
}

func NewServiceStatusRepository(client *ent.Client) user.ServiceStatusRepository {
	return user.NewEntServiceStatusRepository(client)
}

func NewFileRepository(client *ent.Client) file.Repository {
	return file.NewEntRepository(client)
}

func NewFileInfoRepository(client *ent.Client) file.FileInfoRepository {
	return file.NewEntFileInfoRepository(client)
}

func NewFilePathRepository(client *ent.Client) file.FilePathRepository {
	return file.NewEntFilePathRepository(client)
}

func NewFileRoleRepository(client *ent.Client) file.FileRoleRepository {
	return file.NewEntFileRoleRepository(client)
}

// Service constructors

func NewUserService(userRepo user.Repository, statusRepo user.ServiceStatusRepository) user.Service {
	return user.NewService(userRepo, statusRepo)
}

func NewFileService(
	fileRepo file.Repository,
	infoRepo file.FileInfoRepository,
	pathRepo file.FilePathRepository,
	roleRepo file.FileRoleRepository,
) file.Service {
	return file.NewService(fileRepo, infoRepo, pathRepo, roleRepo)
}

func NewStorage(cfg *config.Config) (storage.Storage, error) {
	return s3storage.New(&cfg.Storage)
}

func NewUploadService(fileSvc file.Service, storage storage.Storage) upload.Service {
	return upload.NewService(fileSvc, storage)
}

func NewHandler(
	userSvc user.Service,
	fileSvc file.Service,
	uploadSvc upload.Service,
	cfg *config.Config,
) *v1.Handler {
	return v1.NewHandler(userSvc, fileSvc, uploadSvc, cfg)
}

// Valkey constructor

func NewValkeyClient(cfg *config.Config) (valkey.Client, error) {
	return valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.Cache.Valkey.Addr},
		SelectDB:    cfg.Cache.Valkey.DB,
		Password:    cfg.Cache.Valkey.Password,
	})
}

// Session constructors

func NewSessionRepository(client valkey.Client, cfg *config.Config) auth.SessionRepository {
	return auth.NewValkeySessionRepository(client, cfg.Auth.Session.RedisKey)
}

func NewSessionService(repo auth.SessionRepository, cfg *config.Config) auth.SessionService {
	ttl := time.Duration(cfg.Auth.Session.TTL) * time.Second
	return auth.NewSessionService(repo, ttl)
}

// Keycloak constructor

func NewKeycloak(cfg *config.Config) *auth.Keycloak {
	// Construct issuer URL from Keycloak config
	issuerURL := cfg.Auth.Keycloak.BaseURL + "/realms/" + cfg.Auth.Keycloak.Realm

	// Construct redirect URI from HTTP config
	redirectURI := "http://localhost:8080/auth/callback" // TODO: Make configurable

	return auth.NewKeycloak(auth.KeycloakConfig{
		Issuer:       issuerURL,
		ClientID:     cfg.Auth.Keycloak.ClientID,
		ClientSecret: cfg.Auth.Keycloak.ClientSecret,
		RedirectURI:  redirectURI,
	})
}

func NewOIDCUserService(userSvc user.Service) auth.UserService {
	return auth.NewUserAdapter(userSvc, "keycloak")
}

func NewOIDCService(
	keycloak *auth.Keycloak,
	sessionSvc auth.SessionService,
	userSvc auth.UserService,
	client valkey.Client,
	cfg *config.Config,
) (auth.Service, error) {
	stateTTL := time.Duration(cfg.Auth.Session.StateTTL) * time.Second
	return auth.NewService(keycloak, sessionSvc, userSvc, client, cfg.Auth.Session.RedisKey, stateTTL)
}

func NewAuthHandler(oidcSvc auth.Service, sessionSvc auth.SessionService, cfg *config.Config) *handler.AuthHandler {
	return handler.NewAuthHandler(&handler.AuthHandlerConfig{
		AuthService:    oidcSvc,
		SessionService: sessionSvc,
		Secure:         cfg.Auth.Session.Secure,
	})
}
