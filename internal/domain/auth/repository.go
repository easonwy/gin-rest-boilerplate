package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuthRepository defines the interface for authentication data access
type AuthRepository interface {
	// UserID -> RefreshToken mapping
	SetUserRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiration time.Duration) error
	GetUserRefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteUserRefreshToken(ctx context.Context, userID uuid.UUID) error

	// RefreshToken -> UserID mapping
	SetRefreshTokenUserID(ctx context.Context, token string, userID uuid.UUID, expiration time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (uuid.UUID, error)
	DeleteRefreshTokenUserID(ctx context.Context, token string) error
}
