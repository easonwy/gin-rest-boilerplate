package grpc

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/yi-tech/go-user-service/internal/user/service"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	userpb "github.com/yi-tech/go-user-service/api/proto/user/v1"
	usergrpc "github.com/yi-tech/go-user-service/internal/grpc/user/v1"
)

// Config holds the gRPC server configuration
type Config struct {
	GRPCPort int
	HTTPPort int
}

// Server represents the gRPC server
type Server struct {
	grpcServer *grpc.Server
	httpServer *http.Server
	config     *Config
}

// NewServer creates a new gRPC server
func NewServer(userService service.UserService, config *Config) *Server {
	// Create a gRPC server
	grpcServer := grpc.NewServer()

	// Register the user service
	userServer := usergrpc.NewUserServer(userService)
	userpb.RegisterUserServiceServer(grpcServer, userServer)

	// Enable reflection for gRPC CLI tools
	reflection.Register(grpcServer)

	// Create a new ServeMux for the gRPC-Gateway
	mux := runtime.NewServeMux()

	// Register the gRPC-Gateway handlers
	err := userpb.RegisterUserServiceHandlerFromEndpoint(
		context.Background(),
		mux,
		fmt.Sprintf(":%d", config.GRPCPort),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)
	if err != nil {
		panic(fmt.Sprintf("failed to register gRPC-Gateway: %v", err))
	}

	return &Server{
		grpcServer: grpcServer,
		config:     config,
	}
}

// Start starts the gRPC server
func (s *Server) Start() error {
	// Create a listener on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	// Start gRPC server in a goroutine
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			panic(fmt.Sprintf("failed to serve gRPC: %v", err))
		}
	}()

	// Create a client connection to the gRPC server we just started
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("0.0.0.0:%d", s.config.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to dial server: %v", err)
	}

	// Create a new ServeMux for the HTTP server
	mux := runtime.NewServeMux()

	// Register gRPC Gateway
	if err := userpb.RegisterUserServiceHandler(context.Background(), mux, conn); err != nil {
		return fmt.Errorf("failed to register gateway: %v", err)
	}

	// Create HTTP server with the ServeMux
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler: grpcHandlerFunc(s.grpcServer, mux),
	}

	// Start HTTP server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			panic(fmt.Sprintf("failed to serve HTTP: %v", err))
		}
	}()

	return nil
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() error {
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("failed to shutdown HTTP server: %v", err)
		}
	}

	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	return nil
}

// Config returns the server configuration
func (s *Server) Config() *Config {
	return s.config
}
