package provider

import (
	"fmt"

	"github.com/yi-tech/go-user-service/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerProvider defines methods for providing logger instances
type LoggerProvider interface {
	GetLogger() (*zap.Logger, error)
}

// ZapLoggerProvider implements LoggerProvider using Zap
type ZapLoggerProvider struct {
	cfg *config.Config
}

// NewLoggerProvider creates a new instance of ZapLoggerProvider
func NewLoggerProvider(cfg *config.Config) LoggerProvider {
	return &ZapLoggerProvider{
		cfg: cfg,
	}
}

// GetLogger creates and returns a configured logger instance
func (p *ZapLoggerProvider) GetLogger() (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	// Configure logger based on environment
	if p.cfg.App.Env == "production" {
		// Production configuration with JSON output
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logger, err = config.Build()
	} else {
		// Development configuration with console output
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return logger, nil
}

// Note: The actual Wire provider function is in provider.go
