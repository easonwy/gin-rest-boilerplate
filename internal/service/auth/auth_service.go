package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	// "golang.org/x/crypto/bcrypt" // No longer used directly

	"github.com/yi-tech/go-user-service/internal/config"
	"strings" // Added for strings.Contains

	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
	userService "github.com/yi-tech/go-user-service/internal/service/user" // For user.ErrUserNotFound
)

// Service implements the domainAuth.AuthService interface
type Service struct {
	userService domainUser.UserService
	authRepo    domainAuth.AuthRepository
	config      *config.Config
}

// NewService creates a new auth service instance
func NewService(userService domainUser.UserService, authRepo domainAuth.AuthRepository, config *config.Config) domainAuth.AuthService {
	return &Service{
		userService: userService,
		authRepo:    authRepo,
		config:      config,
	}
}

// Login handles user authentication and token generation
func (s *Service) Login(ctx context.Context, email, password string) (*domainAuth.TokenPair, error) {
	// Find user by email
	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, userService.ErrUserNotFound) {
			return nil, ErrInvalidCredentials // User not found by email
		}
		// For other errors from GetByEmail
		return nil, fmt.Errorf("error retrieving user by email for login: %w", err)
	}
	// If we reach here, user should not be nil if GetByEmail contract is (*User, ErrUserNotFound) or (*User, nil)
	// Adding a safeguard, though ideally GetByEmail guarantees non-nil user if err is nil.
	if user == nil {
	    return nil, ErrInvalidCredentials // Should be unreachable if GetByEmail is consistent
	}

	// Verify password
	if !user.CheckPassword(password) {
		return nil, ErrInvalidCredentials // Password incorrect
	}

	// Generate JWT access token
	expiresAt := time.Now().Add(time.Minute * time.Duration(s.config.JWT.AccessTokenExpireMinutes))
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	})

	accessToken, err := claims.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token and store in repository
	refreshToken := uuid.New().String()
	refreshTokenExpiry := time.Duration(s.config.JWT.RefreshTokenExpireDays) * 24 * time.Hour

	err = s.authRepo.SetUserRefreshToken(ctx, user.ID, refreshToken, refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store user refresh token: %w", err)
	}
	err = s.authRepo.SetRefreshTokenUserID(ctx, refreshToken, user.ID, refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Return token pair
	return &domainAuth.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken handles token refresh logic
func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*domainAuth.TokenPair, error) {
	// Get user ID from the refresh token
	userID, err := s.authRepo.GetUserIDByRefreshToken(ctx, refreshToken) // userID is now uuid.UUID
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrInvalidOrExpiredToken
		}
		return nil, fmt.Errorf("failed to get user ID from refresh token: %w", err)
	}

	// Get user details
	user, err := s.userService.GetByID(ctx, userID) // Pass uuid.UUID directly
	if err != nil {
		if errors.Is(err, userService.ErrUserNotFound) {
			// If user associated with refresh token is not found, token is effectively invalid
			return nil, ErrInvalidOrExpiredToken
		}
		return nil, fmt.Errorf("failed to get user by ID for refresh token: %w", err)
	}

	// Generate new JWT access token
	expiresAt := time.Now().Add(time.Minute * time.Duration(s.config.JWT.AccessTokenExpireMinutes))
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	})

	newAccessToken, err := claims.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign new access token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken := uuid.New().String()
	refreshTokenExpiry := time.Duration(s.config.JWT.RefreshTokenExpireDays) * 24 * time.Hour

	// Store new refresh token
	err = s.authRepo.SetUserRefreshToken(ctx, userID, newRefreshToken, refreshTokenExpiry) // userID is uuid.UUID
	if err != nil {
		return nil, fmt.Errorf("failed to store new user refresh token: %w", err)
	}
	err = s.authRepo.SetRefreshTokenUserID(ctx, newRefreshToken, userID, refreshTokenExpiry) // userID is uuid.UUID
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Delete old refresh token
	err = s.authRepo.DeleteRefreshTokenUserID(ctx, refreshToken)
	if err != nil {
		// Log this error but don't fail the whole operation, as the new token is already set
		fmt.Printf("failed to delete old refresh token to user ID mapping: %v\n", err)
	}

	// Return new token pair
	return &domainAuth.TokenPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout invalidates a user session
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error { // userID is uuid.UUID
	// Get current refresh token for the user
	refreshToken, err := s.authRepo.GetUserRefreshToken(ctx, userID)
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get refresh token during logout: %w", err)
	}

	// Delete refresh token mappings
	if refreshToken != "" {
		err = s.authRepo.DeleteRefreshTokenUserID(ctx, refreshToken)
		if err != nil {
			fmt.Printf("failed to delete refresh token mapping during logout: %v\n", err)
		}
	}

	// Delete user refresh token
	err = s.authRepo.DeleteUserRefreshToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user refresh token during logout: %w", err)
	}

	return nil
}

// ValidateToken validates a JWT token and returns the user ID if valid
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) { // Return uuid.UUID
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return the secret key used for signing
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		// Check for specific JWT errors that indicate an invalid token
		// WORKAROUND: Due to issues resolving jwt.ValidationError constants in the build environment,
		// we are checking error strings. This is fragile and should be replaced with constant checks
		// if the environment/dependency issue is resolved.
		// Common error messages from dgrijalva/jwt-go:
		// "token is expired", "token is not valid yet", "token is malformed", "signature is invalid"

		if strings.Contains(err.Error(), "token is malformed") {
			return uuid.Nil, ErrInvalidToken
		}
		if strings.Contains(err.Error(), "token is expired") {
			return uuid.Nil, ErrInvalidToken // Or a more specific "expired token" error
		}
		if strings.Contains(err.Error(), "token is not valid yet") {
			return uuid.Nil, ErrInvalidToken // Or a more specific "token not yet valid" error
		}
		if strings.Contains(err.Error(), "signature is invalid") {
			return uuid.Nil, ErrInvalidToken
		}
		// The following block is removed due to persistent 'undefined: jwt.ValidationError'
		// var jwtErr *jwt.ValidationError
		// if errors.As(err, &jwtErr) {
		// // If it's a ValidationError, but not caught by specific string checks above,
		// // treat as generic invalid token. The constants are problematic in this env.
		// return uuid.Nil, ErrInvalidToken
		// }
		// If none of the specific string checks caught the error, it might be another type of JWT error or a non-JWT error.
		// We'll rely on the fact that if token.Valid is false later, it will be caught.
		// For errors during parsing not caught by string checks, we'll return a generic parse error.
		// This makes the string checks the primary filter for known JWT issue types.
		return uuid.Nil, fmt.Errorf("failed to parse token (unhandled type or non-JWT error): %w", err)
	}

	// Validate the token
	if !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, ErrInvalidToken // Invalid claims structure
	}

	// Extract user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidToken // user_id claim missing or not a string
	}

	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, ErrInvalidToken // user_id claim is not a valid UUID
	}

	return parsedUserID, nil
}
