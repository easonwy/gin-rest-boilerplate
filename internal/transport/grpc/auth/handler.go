package auth

import (
	"go.uber.org/zap"

	authpb "github.com/yi-tech/go-user-service/api/proto/auth/v1"
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
)

// Handler is a wrapper for the AuthServer to match the wire.go expectations
type Handler struct {
	*AuthServer
}

// NewHandler creates a new auth gRPC handler
func NewHandler(authService domainAuth.AuthService, logger *zap.Logger) *Handler {
	return &Handler{
		AuthServer: NewAuthServer(authService, logger),
	}
}

// GetServer returns the underlying AuthServer for registration with gRPC
func (h *Handler) GetServer() authpb.AuthServiceServer {
	return h.AuthServer
}
