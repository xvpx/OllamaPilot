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
	DBType     string `env:"DB_TYPE" envDefault:"postgres"`
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBName     string `env:"DB_NAME" envDefault:"ollamapilot"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:""`
	DBSSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`
	
	// Legacy SQLite support
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
	
	// Semantic Memory configuration
	EnableSemanticMemory bool   `env:"ENABLE_SEMANTIC_MEMORY" envDefault:"true"`
	EmbeddingModel       string `env:"EMBEDDING_MODEL" envDefault:"nomic-embed-text"`
	MaxContextResults    int    `env:"MAX_CONTEXT_RESULTS" envDefault:"5"`
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

	if c.DBType == "postgres" {
		if c.DBHost == "" {
			return fmt.Errorf("DB_HOST cannot be empty for PostgreSQL")
		}
		if c.DBName == "" {
			return fmt.Errorf("DB_NAME cannot be empty for PostgreSQL")
		}
		if c.DBUser == "" {
			return fmt.Errorf("DB_USER cannot be empty for PostgreSQL")
		}
	} else if c.DBType == "sqlite" {
		if c.DBPath == "" {
			return fmt.Errorf("DB_PATH cannot be empty for SQLite")
		}
	} else {
		return fmt.Errorf("unsupported DB_TYPE: %s (supported: postgres, sqlite)", c.DBType)
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

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	if c.DBType == "postgres" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
	}
	// Default to SQLite for backward compatibility
	return c.DBPath
}

// IsPostgreSQL returns true if using PostgreSQL
func (c *Config) IsPostgreSQL() bool {
	return c.DBType == "postgres"
}

// IsSQLite returns true if using SQLite
func (c *Config) IsSQLite() bool {
	return c.DBType == "sqlite"
}