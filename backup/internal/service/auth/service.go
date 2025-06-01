package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/config"
	"github.com/yi-tech/go-user-service/internal/domain/auth"
	"github.com/yi-tech/go-user-service/internal/domain/user"
	"go.uber.org/zap"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrSessionNotFound    = errors.New("session not found")
)

// Claims represents the JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.StandardClaims
}

// authService implements the domain auth.Service interface
type authService struct {
	userService user.Service
	authRepo    auth.Repository
	config      *config.Config
	logger      *zap.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(userService user.Service, authRepo auth.Repository, config *config.Config, logger *zap.Logger) auth.Service {
	return &authService{
		userService: userService,
		authRepo:    authRepo,
		config:      config,
		logger:      logger,
	}
}

// Login authenticates a user and creates a session
func (s *authService) Login(ctx context.Context, email, password, userAgent, clientIP string) (*auth.TokenPair, error) {
	// Get user by email
	user, err := s.userService.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error("Failed to get user during login", zap.Error(err), zap.String("email", email))
		return nil, ErrInvalidCredentials
	}

	// Validate password
	if !user.ValidatePassword(password) {
		s.logger.Warn("Invalid password attempt", zap.String("email", email), zap.String("ip", clientIP))
		return nil, ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := s.generateTokenPair(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Create session
	session := auth.NewSession(
		user.ID,
		tokenPair.RefreshToken,
		userAgent,
		clientIP,
		time.Duration(s.config.JWT.RefreshTokenExpireDays)*24*time.Hour,
	)

	// Save session
	if err := s.authRepo.CreateSession(ctx, session); err != nil {
		s.logger.Error("Failed to save session", zap.Error(err))
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return tokenPair, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken, userAgent, clientIP string) (*auth.TokenPair, error) {
	// Get session by refresh token
	session, err := s.authRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Error("Failed to get session by refresh token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Check if session is expired
	if session.IsExpired() {
		s.logger.Info("Attempted to use expired refresh token", zap.String("session_id", session.ID))
		// Delete expired session
		_ = s.authRepo.DeleteSession(ctx, session.ID)
		return nil, ErrInvalidToken
	}

	// Generate new token pair
	tokenPair, err := s.generateTokenPair(session.UserID)
	if err != nil {
		s.logger.Error("Failed to generate new tokens", zap.Error(err))
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Delete old session
	if err := s.authRepo.DeleteSession(ctx, session.ID); err != nil {
		s.logger.Error("Failed to delete old session", zap.Error(err), zap.String("session_id", session.ID))
	}

	// Create new session
	newSession := auth.NewSession(
		session.UserID,
		tokenPair.RefreshToken,
		userAgent,
		clientIP,
		time.Duration(s.config.JWT.RefreshTokenExpireDays)*24*time.Hour,
	)

	// Save new session
	if err := s.authRepo.CreateSession(ctx, newSession); err != nil {
		s.logger.Error("Failed to save new session", zap.Error(err))
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	return tokenPair, nil
}

// Logout invalidates a session
func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	// Get session by refresh token
	session, err := s.authRepo.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		s.logger.Error("Failed to get session during logout", zap.Error(err))
		return ErrInvalidToken
	}

	if session == nil {
		return ErrSessionNotFound
	}

	// Delete session
	if err := s.authRepo.DeleteSession(ctx, session.ID); err != nil {
		s.logger.Error("Failed to delete session", zap.Error(err), zap.String("session_id", session.ID))
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// ValidateToken validates an access token and returns the user ID
func (s *authService) ValidateToken(ctx context.Context, accessToken string) (uuid.UUID, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWT.Secret), nil
	})

	if err != nil {
		s.logger.Error("Failed to parse token", zap.Error(err))
		return uuid.Nil, ErrInvalidToken
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return uuid.Nil, ErrInvalidToken
}

// GetUserFromToken retrieves a user from an access token
func (s *authService) GetUserFromToken(ctx context.Context, accessToken string) (*user.User, error) {
	// Validate token and get user ID
	userID, err := s.ValidateToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	// Get user by ID
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user from token", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// generateTokenPair generates a new access and refresh token pair
func (s *authService) generateTokenPair(userID uuid.UUID) (*auth.TokenPair, error) {
	// Create access token
	accessTokenExpiry := time.Now().Add(time.Duration(s.config.JWT.AccessTokenExpireMinutes) * time.Minute)
	accessTokenClaims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: accessTokenExpiry.Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   userID.String(),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWT.Secret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token
	refreshToken := uuid.NewString()

	return &auth.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshToken,
	}, nil
}
