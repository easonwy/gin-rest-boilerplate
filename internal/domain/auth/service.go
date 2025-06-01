package auth

import (
	"context"

	"github.com/yi-tech/go-user-service/internal/domain/auth/dto"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	// Login authenticates a user and creates a session
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
	
	// RefreshToken refreshes an access token using a refresh token
	RefreshToken(refreshToken string) (*dto.LoginResponse, error)
	
	// Logout invalidates a session
	Logout(ctx context.Context, userID uint) error
	
	// ValidateToken validates an access token and returns the user ID
	ValidateToken(ctx context.Context, accessToken string) (uint, error)
}
