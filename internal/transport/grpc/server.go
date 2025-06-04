package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/yi-tech/go-user-service/api/proto/auth/v1"
	userpb "github.com/yi-tech/go-user-service/api/proto/user/v1"
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
	serviceUser "github.com/yi-tech/go-user-service/internal/service/user"
	grpcAuth "github.com/yi-tech/go-user-service/internal/transport/grpc/auth"
	grpcUser "github.com/yi-tech/go-user-service/internal/transport/grpc/user"
)

// Config represents the gRPC server configuration
type Config struct {
	GRPCPort int
	HTTPPort int
}

// Server represents the gRPC server
type Server struct {
	userHandler *grpcUser.Handler
	authHandler *grpcAuth.Handler
	logger      *zap.Logger
	cfg         *Config
	server      *grpc.Server
	httpServer  *http.Server
}

// NewServer creates a new gRPC server
func NewServer(userService serviceUser.UserService, authService domainAuth.AuthService, logger *zap.Logger, cfg *Config) *Server {
	return &Server{
		userHandler: grpcUser.NewHandler(userService, logger),
		authHandler: grpcAuth.NewHandler(authService, logger),
		logger:      logger,
		cfg:         cfg,
	}
}

// Start starts the gRPC server and the HTTP gateway
func (s *Server) Start() error {
	// Create a listener for the gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Create a new gRPC server
	s.server = grpc.NewServer()

	// Register services
	authpb.RegisterAuthServiceServer(s.server, s.authHandler.GetServer())
	userpb.RegisterUserServiceServer(s.server, s.userHandler.GetServer())

	// Start the gRPC server in a goroutine
	go func() {
		s.logger.Info("Starting gRPC server", zap.Int("port", s.cfg.GRPCPort))
		if err := s.server.Serve(lis); err != nil {
			s.logger.Error("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Start the HTTP gateway
	return s.startHTTPGateway()
}

// startHTTPGateway starts the HTTP gateway for the gRPC server
func (s *Server) startHTTPGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a new mux for the HTTP gateway
	mux := runtime.NewServeMux()

	// Set up a connection to the gRPC server
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	grpcServerEndpoint := fmt.Sprintf("localhost:%d", s.cfg.GRPCPort)

	// Register services
	err := authpb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register auth service handler: %v", err)
	}

	err = userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts)
	if err != nil {
		return fmt.Errorf("failed to register user service handler: %v", err)
	}

	// Create a new HTTP server for the gateway
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.HTTPPort),
		Handler: mux,
	}

	// Start the HTTP server in a goroutine
	go func() {
		s.logger.Info("Starting HTTP gateway", zap.Int("port", s.cfg.HTTPPort))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Failed to serve HTTP gateway", zap.Error(err))
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the gRPC server and the HTTP gateway
func (s *Server) Shutdown(ctx context.Context) error {
	// Shutdown the HTTP gateway
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown HTTP gateway", zap.Error(err))
		}
	}

	// Gracefully stop the gRPC server
	if s.server != nil {
		s.server.GracefulStop()
	}
	return nil
}

// Stop gracefully shuts down the gRPC server and the HTTP gateway
func (s *Server) Stop() error {
	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return s.Shutdown(ctx)
}
