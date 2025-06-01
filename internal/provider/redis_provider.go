package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yi-tech/go-user-service/internal/config"
)

// RedisProvider defines methods for providing Redis client connections
type RedisProvider interface {
	GetRedisClient() (*redis.Client, error)
}

// DefaultRedisProvider implements RedisProvider
type DefaultRedisProvider struct {
	cfg *config.Config
}

// NewRedisProvider creates a new instance of DefaultRedisProvider
func NewRedisProvider(cfg *config.Config) RedisProvider {
	return &DefaultRedisProvider{
		cfg: cfg,
	}
}

// GetRedisClient creates and returns a configured Redis client
func (p *DefaultRedisProvider) GetRedisClient() (*redis.Client, error) {
	// Create Redis client with configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:         p.cfg.Redis.Addr,
		Password:     p.cfg.Redis.Password,
		DB:           p.cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Ping the Redis server to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return rdb, nil
}

// Note: The actual Wire provider function is in provider.go
