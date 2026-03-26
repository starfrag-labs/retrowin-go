package retrowinserver

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	apiv1 "github.com/starfrag-lab/retrowin-go/pkg/api/v1"

	"github.com/starfrag-lab/retrowin-go/internal/config"
	"github.com/starfrag-lab/retrowin-go/internal/handler"
	"github.com/starfrag-lab/retrowin-go/internal/handler/v1"
)

// Server represents the HTTP server.
type Server struct {
	httpServer  *http.Server
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
	// Create ogen server
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

// Start starts the HTTP server.
func (s *Server) Start(lc fx.Lifecycle) error {
	addr := fmt.Sprintf("%s:%d", s.config.HTTP.Host, s.config.HTTP.Port)

	// Create a mux to combine API and auth routes
	mux := http.NewServeMux()

	// Mount auth routes
	mux.HandleFunc("/auth/login", s.authHandler.Login)
	mux.HandleFunc("/auth/callback", s.authHandler.Callback)
	mux.HandleFunc("/auth/logout", s.authHandler.Logout)
	mux.HandleFunc("/auth/me", s.authHandler.GetMe)

	// Mount API routes (everything else goes to ogen)
	mux.Handle("/", s.ogenServer)

	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.logger.Info("starting HTTP server", zap.String("addr", addr))
			go func() {
				if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.logger.Error("HTTP server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.logger.Info("stopping HTTP server")
			return s.httpServer.Shutdown(ctx)
		},
	})

	return nil
}
