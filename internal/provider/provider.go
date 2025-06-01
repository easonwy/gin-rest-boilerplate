package provider

// This file contains exported provider functions for Wire dependency injection.
// It acts as a facade for the actual provider implementations.

import (
	"github.com/go-redis/redis/v8"
	"github.com/yi-tech/go-user-service/internal/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProvideConfig is the Wire provider function for application configuration.
// It delegates to the implementation in config_provider.go.
func ProvideConfig() (*config.Config, error) {
	provider := NewConfigProvider()
	return provider.GetConfig()
}

// ProvideLogger is the Wire provider function for the logger.
// It delegates to the implementation in logger_provider.go.
func ProvideLogger(cfg *config.Config) (*zap.Logger, error) {
	provider := NewLoggerProvider(cfg)
	return provider.GetLogger()
}

// ProvideDatabase is the Wire provider function for the database connection.
// It delegates to the implementation in database_provider.go.
func ProvideDatabase(cfg *config.Config) (*gorm.DB, error) {
	provider := NewDatabaseProvider(cfg)
	return provider.GetDB()
}

// ProvideRedisClient is the Wire provider function for the Redis client.
// It delegates to the implementation in redis_provider.go.
func ProvideRedisClient(cfg *config.Config) (*redis.Client, error) {
	provider := NewRedisProvider(cfg)
	return provider.GetRedisClient()
}
