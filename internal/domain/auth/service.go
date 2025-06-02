package auth

import (
	"context"

	"github.com/google/uuid"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	// Login authenticates a user and returns a token pair
	Login(ctx context.Context, email, password string) (*TokenPair, error)

	// RefreshToken refreshes an access token using a refresh token and returns a new token pair
	RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

	// Logout invalidates a session
	Logout(ctx context.Context, userID uuid.UUID) error

	// ValidateToken validates an access token and returns the user ID
	ValidateToken(ctx context.Context, accessToken string) (uuid.UUID, error)
}
