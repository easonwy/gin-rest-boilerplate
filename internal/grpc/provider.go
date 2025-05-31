// Package grpc provides gRPC server implementation
package grpc

import (
	"github.com/google/wire"
	"github.com/yi-tech/go-user-service/internal/config"
	"github.com/yi-tech/go-user-service/internal/user/repository"
	userservice "github.com/yi-tech/go-user-service/internal/user/service"
)

// ProviderSet is a wire provider set for gRPC server
var ProviderSet = wire.NewSet(
	NewServer,
	NewConfig,
	userservice.NewUserService,
	repository.NewUserRepository,
)

// NewConfig creates a new gRPC server configuration from the app config
func NewConfig(cfg *config.Config) *Config {
	return &Config{
		GRPCPort: cfg.GRPC.Port,
		HTTPPort: 8080, // This is the gRPC-Gateway port, not the main HTTP port
	}
}
