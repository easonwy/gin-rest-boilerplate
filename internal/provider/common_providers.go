package provider

import (
	"context"
	"log"
	"time"

	"github.com/example/go-user-service/internal/config"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProvideConfig provides the application configuration.
func ProvideConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
		return nil, err // Should not reach here due to log.Fatalf, but for signature compliance
	}
	return cfg, nil
}

// ProvideDatabase provides a GORM database connection.
func ProvideDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.Source), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return nil, err // Should not reach here due to log.Fatalf, but for signature compliance
	}

	// Auto-migrate the User model (This might be moved elsewhere in a larger app)
	// err = db.AutoMigrate(&model.User{})
	// if err != nil {
	// 	log.Fatalf("Failed to auto-migrate database: %v", err)
	// 	return nil, err
	// }

	return db, nil
}

// ProvideRedisClient provides a Redis client connection.
func ProvideRedisClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Ping the Redis server to check the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
		return nil, err // Should not reach here due to log.Fatalf, but for signature compliance
	}

	return rdb, nil
}
