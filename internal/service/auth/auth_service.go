package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/config"
	"github.com/yi-tech/go-user-service/internal/domain/auth"
	"github.com/yi-tech/go-user-service/internal/domain/auth/dto"
	userService "github.com/yi-tech/go-user-service/internal/service/user"
	"golang.org/x/crypto/bcrypt"
)

// Service implements the auth.AuthService interface
type Service struct {
	userService userService.UserService
	authRepo    auth.AuthRepository
	config      *config.Config
}

// NewService creates a new auth service instance
func NewService(userService userService.UserService, authRepo auth.AuthRepository, config *config.Config) auth.AuthService {
	return &Service{
		userService: userService,
		authRepo:    authRepo,
		config:      config,
	}
}

// Login handles user authentication and token generation
func (s *Service) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	// Find user by email
	ctx := context.Background()
	user, err := s.userService.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to compare password hash: %w", err)
	}

	// Generate JWT access token
	expiresAt := time.Now().Add(time.Minute * time.Duration(s.config.JWT.AccessTokenExpireMinutes))
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
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

	err = s.authRepo.SetUserRefreshToken(context.Background(), user.ID, refreshToken, refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store user refresh token: %w", err)
	}
	err = s.authRepo.SetRefreshTokenUserID(context.Background(), refreshToken, user.ID, refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Return login response with tokens
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresAt.Unix() - time.Now().Unix(),
	}, nil
}

// RefreshToken handles token refresh logic
func (s *Service) RefreshToken(refreshToken string) (*dto.LoginResponse, error) {
	// Get user ID from the refresh token
	userID, err := s.authRepo.GetUserIDByRefreshToken(context.Background(), refreshToken)
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("failed to get user ID from refresh token: %w", err)
	}

	// Get user details
	ctx := context.Background()
	userUUID, err := uuid.Parse(fmt.Sprintf("%d", userID))
	if err != nil {
		return nil, fmt.Errorf("failed to parse user ID: %w", err)
	}
	user, err := s.userService.GetByID(ctx, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// Generate new JWT access token
	expiresAt := time.Now().Add(time.Minute * time.Duration(s.config.JWT.AccessTokenExpireMinutes))
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
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
	err = s.authRepo.SetUserRefreshToken(context.Background(), uint(userID), newRefreshToken, refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store new user refresh token: %w", err)
	}
	err = s.authRepo.SetRefreshTokenUserID(context.Background(), newRefreshToken, uint(userID), refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// Delete old refresh token
	err = s.authRepo.DeleteRefreshTokenUserID(context.Background(), refreshToken)
	if err != nil {
		fmt.Printf("failed to delete old refresh token to user ID mapping: %v\n", err)
	}

	// Return new tokens
	return &dto.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresAt.Unix() - time.Now().Unix(),
	}, nil
}

// Logout invalidates a user session
func (s *Service) Logout(ctx context.Context, userID uint) error {
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
func (s *Service) ValidateToken(ctx context.Context, tokenString string) (uint, error) {
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
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	// Validate the token
	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	// Extract user ID from claims
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user ID in token")
	}

	userID := uint(userIDFloat)
	return userID, nil
}
