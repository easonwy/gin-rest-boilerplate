package auth

import (
	"context"
	"time"
)

// AuthRepository defines the interface for authentication data access
type AuthRepository interface {
	// UserID -> RefreshToken mapping
	SetUserRefreshToken(ctx context.Context, userID uint, token string, expiration time.Duration) error
	GetUserRefreshToken(ctx context.Context, userID uint) (string, error)
	DeleteUserRefreshToken(ctx context.Context, userID uint) error

	// RefreshToken -> UserID mapping
	SetRefreshTokenUserID(ctx context.Context, token string, userID uint, expiration time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, token string) (uint, error)
	DeleteRefreshTokenUserID(ctx context.Context, token string) error
}
