package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v10"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port        string `env:"PORT" envDefault:"8080"`
	Host        string `env:"HOST" envDefault:"0.0.0.0"`
	Environment string `env:"ENV" envDefault:"development"`

	// Database configuration
	DBPath string `env:"DB_PATH" envDefault:"./data/chat.db"`

	// Ollama configuration
	OllamaHost    string        `env:"OLLAMA_HOST" envDefault:"localhost:11434"`
	OllamaTimeout time.Duration `env:"OLLAMA_TIMEOUT" envDefault:"30s"`

	// Logging configuration
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`

	// Performance configuration
	MaxConcurrentChats int           `env:"MAX_CONCURRENT_CHATS" envDefault:"100"`
	ReadTimeout        time.Duration `env:"READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout       time.Duration `env:"WRITE_TIMEOUT" envDefault:"30s"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Load from environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Port == "" {
		return fmt.Errorf("PORT cannot be empty")
	}

	if c.DBPath == "" {
		return fmt.Errorf("DB_PATH cannot be empty")
	}

	if c.OllamaHost == "" {
		return fmt.Errorf("OLLAMA_HOST cannot be empty")
	}

	if c.MaxConcurrentChats <= 0 {
		return fmt.Errorf("MAX_CONCURRENT_CHATS must be positive")
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}