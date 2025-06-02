package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/yi-tech/go-user-service/internal/config"
	domainAuth "github.com/yi-tech/go-user-service/internal/domain/auth"
)

// authRepository struct implements the domainAuth.AuthRepository interface
type AuthRepositoryImpl struct {
	redisClient *redis.Client
}

// NewAuthRepository creates a new instance of AuthRepository.
func NewAuthRepository(redisClient *redis.Client) domainAuth.AuthRepository { // Return type changed to domain interface
	return &AuthRepositoryImpl{redisClient: redisClient}
}

func (r *AuthRepositoryImpl) SetUserRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiration time.Duration) error { // userID type changed
	key := fmt.Sprintf(config.RedisKeyPrefix+"refresh_token:%s", userID.String()) // key formatting changed
	err := r.redisClient.Set(ctx, key, token, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set refresh token in redis: %w", err)
	}
	return nil
}

func (r *AuthRepositoryImpl) GetUserRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) { // userID type changed
	key := fmt.Sprintf(config.RedisKeyPrefix+"refresh_token:%s", userID.String()) // key formatting changed
	token, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Token not found, service layer should handle this
		}
		return "", fmt.Errorf("failed to get refresh token from redis: %w", err)
	}
	return token, nil
}

func (r *AuthRepositoryImpl) DeleteUserRefreshToken(ctx context.Context, userID uuid.UUID) error { // userID type changed
	key := fmt.Sprintf(config.RedisKeyPrefix+"refresh_token:%s", userID.String()) // key formatting changed
	err := r.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete refresh token from redis: %w", err)
	}
	return nil
}

func (r *AuthRepositoryImpl) SetRefreshTokenUserID(ctx context.Context, token string, userID uuid.UUID, expiration time.Duration) error { // userID type changed
	key := fmt.Sprintf(config.RedisKeyPrefix+"user_id:%s", token)
	err := r.redisClient.Set(ctx, key, userID.String(), expiration).Err() // Store userID.String()
	if err != nil {
		return fmt.Errorf("failed to set user ID by refresh token in redis: %w", err)
	}
	return nil
}

func (r *AuthRepositoryImpl) GetUserIDByRefreshToken(ctx context.Context, token string) (uuid.UUID, error) { // return type and userID type changed
	key := fmt.Sprintf(config.RedisKeyPrefix+"user_id:%s", token)
	userIDStr, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return uuid.Nil, nil // User ID not found, service layer should handle this
		}
		return uuid.Nil, fmt.Errorf("failed to get user ID by refresh token from redis: %w", err)
	}
	
	parsedUserID, err := uuid.Parse(userIDStr) // Parse string to uuid.UUID
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user ID '%s' from redis: %w", userIDStr, err)
	}
	return parsedUserID, nil
}

func (r *AuthRepositoryImpl) DeleteRefreshTokenUserID(ctx context.Context, token string) error {
	key := fmt.Sprintf(config.RedisKeyPrefix+"user_id:%s", token)
	err := r.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user ID by refresh token from redis: %w", err)
	}
	return nil
}
