package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tapas/go-user-service/internal/auth/model"
	"github.com/tapas/go-user-service/internal/auth/repository"
	"github.com/tapas/go-user-service/internal/config"
	"github.com/tapas/go-user-service/internal/user/service"
	"golang.org/x/crypto/bcrypt"
)

// AuthService 定义认证服务接口
type AuthService interface {
	Login(req model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(refreshToken string) (*model.LoginResponse, error)
	Logout(ctx context.Context, userID uint) error
}

// authService 认证服务实现结构体
type authService struct {
	userService service.UserService
	authRepo    repository.AuthRepository
	config      *config.Config
}

// NewAuthService 创建新的认证服务实例
func NewAuthService(userService service.UserService, authRepo repository.AuthRepository, config *config.Config) AuthService {
	return &authService{
		userService: userService,
		authRepo:    authRepo,
		config:      config,
	}
}

// Login 处理用户登录逻辑
func (s *authService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	// 1. 通过用户服务查找用户
	user, err := s.userService.GetUserByEmail(req.Email)
	if err != nil {
		// TODO: Handle specific errors like user not found
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// 2. 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.New("invalid credentials") // Password mismatch
		}
		return nil, fmt.Errorf("failed to compare password hash: %w", err)
	}

	// 3. 生成 JWT Access Token
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

	// 4. 生成 Refresh Token 并存储到 Redis (using repository)
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

	// 5. 返回 LoginResponse
	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresAt.Unix() - time.Now().Unix(),
	}, nil
}

// RefreshToken 处理刷新令牌逻辑
func (s *authService) RefreshToken(refreshToken string) (*model.LoginResponse, error) {
	// 1. 从 Redis 获取用户 ID (using repository)
	// The refresh token itself is the key in the repository, not the user ID.
	// We need to get the user ID associated with the refresh token.
	// Let's assume the repository stores refresh token -> user ID mapping.
	// The GetRefreshToken method should return the user ID (as string) or an error.

	// Corrected logic: Get user ID from the refresh token string
	// This requires a change in the AuthRepository interface and implementation
	// to map refresh token to user ID.

	// For now, let's revert to the previous Redis logic to unblock the build
	// and address the repository mapping later.

	userID, err := s.authRepo.GetUserIDByRefreshToken(context.Background(), refreshToken)
	if err != nil {
		if err == redis.Nil {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, fmt.Errorf("failed to get user ID from refresh token: %w", err)
	}

	// 2. 通过用户服务查找用户
	user, err := s.userService.GetUserByID(uint(userID))
	if err != nil {
		// TODO: Handle specific errors like user not found
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	// 3. 生成新的 JWT Access Token
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

	// 4. 生成新的 Refresh Token 并存储到 Redis (using repository)
	newRefreshToken := uuid.New().String()
	refreshTokenExpiry := time.Duration(s.config.JWT.RefreshTokenExpireDays) * 24 * time.Hour

	err = s.authRepo.SetUserRefreshToken(context.Background(), uint(userID), newRefreshToken, refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store new user refresh token: %w", err)
	}
	err = s.authRepo.SetRefreshTokenUserID(context.Background(), newRefreshToken, uint(userID), refreshTokenExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to store new refresh token: %w", err)
	}

	// 5. 删除旧的 Refresh Token (using repository)
	// The repository's DeleteRefreshToken method expects a user ID.
	// We need to delete the refresh token using the token string itself as the key.

	err = s.authRepo.DeleteRefreshTokenUserID(context.Background(), refreshToken)
	if err != nil {
		fmt.Printf("failed to delete old refresh token to user ID mapping: %v\n", err)
	}
	err = s.authRepo.DeleteUserRefreshToken(context.Background(), uint(userID))
	if err != nil {
		fmt.Printf("failed to delete old user refresh token: %v\n", err)
	}

	// 6. 返回新的 LoginResponse
	return &model.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresAt.Unix() - time.Now().Unix(),
	}, nil
}

// Logout handles user logout logic
func (s *authService) Logout(ctx context.Context, userID uint) error {
	// Delete the refresh token associated with the user ID
	err := s.authRepo.DeleteUserRefreshToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token during logout: %w", err)
	}

	return nil
}
