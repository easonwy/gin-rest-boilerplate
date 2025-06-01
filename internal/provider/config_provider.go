package provider

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/yi-tech/go-user-service/internal/config"
)

// ConfigProvider defines the interface for configuration providers
type ConfigProvider interface {
	GetConfig() (*config.Config, error)
}

// DefaultConfigProvider implements ConfigProvider using Viper
type DefaultConfigProvider struct{
	configPath string
}

// NewConfigProvider creates a new instance of DefaultConfigProvider
// By default, it looks for config files in the ./configs directory
func NewConfigProvider() ConfigProvider {
	return &DefaultConfigProvider{
		configPath: "./configs",
	}
}

// NewConfigProviderWithPath creates a new instance of DefaultConfigProvider with a custom config path
func NewConfigProviderWithPath(configPath string) ConfigProvider {
	return &DefaultConfigProvider{
		configPath: configPath,
	}
}

// GetConfig loads and returns the application configuration
func (p *DefaultConfigProvider) GetConfig() (*config.Config, error) {
	// Determine environment
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local" // Default to local environment
	}

	// Initialize Viper
	v := viper.New()
	v.SetConfigName(fmt.Sprintf("config.%s", env))
	v.SetConfigType("yaml")
	v.AddConfigPath(p.configPath) // Look for config in the specified directory

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into Config struct
	var cfg config.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Note: The actual Wire provider function is in provider.go
