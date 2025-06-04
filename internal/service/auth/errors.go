package auth

import "errors"

// Service-level errors for authentication and authorization operations
var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidOrExpiredToken = errors.New("invalid or expired refresh token")
	ErrInvalidToken          = errors.New("invalid token") // For general token validation issues
)
