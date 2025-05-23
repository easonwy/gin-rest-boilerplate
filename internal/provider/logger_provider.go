package provider

import (
	"log"

	"go.uber.org/zap"
)

// ProvideLogger provides a Zap logger instance.
func ProvideLogger() (*zap.Logger, error) {
	// For simplicity, using a development logger. In production, use zap.NewProduction()
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to create Zap logger: %v", err)
		return nil, err
	}

	// Optional: Replace global logger (use with caution)
	// zap.ReplaceGlobals(logger)

	return logger, nil
}
