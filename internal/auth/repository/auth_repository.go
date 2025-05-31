package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yi-tech/go-user-service/internal/config"
)

// AuthRepository defines the interface for authentication data operations.
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

type authRepository struct {
	redisClient *redis.Client
}

// NewAuthRepository creates a new instance of AuthRepository.
func NewAuthRepository(redisClient *redis.Client) AuthRepository {
	return &authRepository{redisClient: redisClient}
}

func (r *authRepository) SetUserRefreshToken(ctx context.Context, userID uint, token string, expiration time.Duration) error {
	// Store refresh token in Redis with user ID as value
	key := fmt.Sprintf(config.RedisKeyPrefix+"refresh_token:%d", userID)
	err := r.redisClient.Set(ctx, key, token, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set refresh token in redis: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserRefreshToken(ctx context.Context, userID uint) (string, error) {
	// Get refresh token from Redis by user ID
	key := fmt.Sprintf(config.RedisKeyPrefix+"refresh_token:%d", userID)
	token, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Token not found
		}
		return "", fmt.Errorf("failed to get refresh token from redis: %w", err)
	}
	return token, nil
}

func (r *authRepository) DeleteUserRefreshToken(ctx context.Context, userID uint) error {
	// Delete refresh token from Redis by user ID
	key := fmt.Sprintf(config.RedisKeyPrefix+"refresh_token:%d", userID)
	err := r.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token from redis: %w", err)
	}
	return nil
}

func (r *authRepository) SetRefreshTokenUserID(ctx context.Context, token string, userID uint, expiration time.Duration) error {
	// Store user ID in Redis with refresh token as key
	key := fmt.Sprintf(config.RedisKeyPrefix+"user_id:%s", token)
	err := r.redisClient.Set(ctx, key, userID, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set user ID by refresh token in redis: %w", err)
	}
	return nil
}

func (r *authRepository) GetUserIDByRefreshToken(ctx context.Context, token string) (uint, error) {
	// Get user ID from Redis by refresh token
	key := fmt.Sprintf(config.RedisKeyPrefix+"user_id:%s", token)
	userIDStr, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // User ID not found
		}
		return 0, fmt.Errorf("failed to get user ID by refresh token from redis: %w", err)
	}
	var userID uint
	_, err = fmt.Sscanf(userIDStr, "%d", &userID)
	if err != nil {
		return 0, fmt.Errorf("failed to parse user ID from redis: %w", err)
	}
	return userID, nil
}

func (r *authRepository) DeleteRefreshTokenUserID(ctx context.Context, token string) error {
	// Delete user ID from Redis by refresh token
	key := fmt.Sprintf(config.RedisKeyPrefix+"user_id:%s", token)
	err := r.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user ID by refresh token from redis: %w", err)
	}
	return nil
}
