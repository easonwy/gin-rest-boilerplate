package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// AuthRepository defines the interface for authentication data operations.
type AuthRepository interface {
	SetRefreshToken(ctx context.Context, userID uint, token string, expiration time.Duration) error
	GetRefreshToken(ctx context.Context, userID uint) (string, error)
	DeleteRefreshToken(ctx context.Context, userID uint) error
}

type authRepository struct {
	redisClient *redis.Client
}

// NewAuthRepository creates a new instance of AuthRepository.
func NewAuthRepository(redisClient *redis.Client) AuthRepository {
	return &authRepository{redisClient: redisClient}
}

func (r *authRepository) SetRefreshToken(ctx context.Context, userID uint, token string, expiration time.Duration) error {
	// Store refresh token in Redis with user ID as value
	key := fmt.Sprintf("refresh_token:%d", userID)
	err := r.redisClient.Set(ctx, key, token, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set refresh token in redis: %w", err)
	}
	return nil
}

func (r *authRepository) GetRefreshToken(ctx context.Context, userID uint) (string, error) {
	// Get refresh token from Redis by user ID
	key := fmt.Sprintf("refresh_token:%d", userID)
	token, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Token not found
		}
		return "", fmt.Errorf("failed to get refresh token from redis: %w", err)
	}
	return token, nil
}

func (r *authRepository) DeleteRefreshToken(ctx context.Context, userID uint) error {
	// Delete refresh token from Redis by user ID
	key := fmt.Sprintf("refresh_token:%d", userID)
	err := r.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token from redis: %w", err)
	}
	return nil
}
