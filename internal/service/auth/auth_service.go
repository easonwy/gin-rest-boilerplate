package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/yi-tech/go-user-service/internal/config"
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
	domainUser "github.com/yi-tech/go-user-service/internal/domain/user"
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
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Check if user exists
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to compare password hash: %w", err)
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
		if err == redis.Nil {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("failed to get user ID from refresh token: %w", err)
	}

	// Get user details
	user, err := s.userService.GetByID(ctx, userID) // Pass uuid.UUID directly
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
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
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Validate the token
	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid token claims")
	}

	// Extract user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("user_id claim is not a string")
	}

	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user_id claim to UUID: %w", err)
	}

	return parsedUserID, nil
}
