package provider

import (
	"fmt"

	"github.com/yi-tech/go-user-service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseProvider defines methods for providing database connections
type DatabaseProvider interface {
	GetDB() (*gorm.DB, error)
}

// GormDatabaseProvider implements DatabaseProvider using GORM
type GormDatabaseProvider struct {
	cfg *config.Config
}

// NewDatabaseProvider creates a new instance of GormDatabaseProvider
func NewDatabaseProvider(cfg *config.Config) DatabaseProvider {
	return &GormDatabaseProvider{
		cfg: cfg,
	}
}

// GetDB creates and returns a configured database connection
func (p *GormDatabaseProvider) GetDB() (*gorm.DB, error) {
	// Configure GORM with more options for production readiness
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// Add other GORM configuration options as needed
	}

	db, err := gorm.Open(postgres.Open(p.cfg.Database.Source), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get the underlying SQL DB to set connection pool parameters
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQL DB: %w", err)
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

// Note: The actual Wire provider function is in provider.go
