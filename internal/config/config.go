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

	// Database configuration (PostgreSQL only)
	DBHost     string `env:"DB_HOST" envDefault:"localhost"`
	DBPort     string `env:"DB_PORT" envDefault:"5432"`
	DBName     string `env:"DB_NAME" envDefault:"ollamapilot"`
	DBUser     string `env:"DB_USER" envDefault:"postgres"`
	DBPassword string `env:"DB_PASSWORD" envDefault:""`
	DBSSLMode  string `env:"DB_SSL_MODE" envDefault:"disable"`

	// Ollama configuration
	OllamaHost    string        `env:"OLLAMA_HOST" envDefault:"localhost:11434"`
	OllamaTimeout time.Duration `env:"OLLAMA_TIMEOUT" envDefault:"300s"`

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
	
	// Authentication configuration
	JWTSecret     string        `env:"JWT_SECRET" envDefault:"your-secret-key-change-in-production"`
	JWTExpiration time.Duration `env:"JWT_EXPIRATION" envDefault:"24h"`
	BCryptCost    int           `env:"BCRYPT_COST" envDefault:"8"`
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

	// Validate PostgreSQL configuration
	if c.DBHost == "" {
		return fmt.Errorf("DB_HOST cannot be empty")
	}
	if c.DBName == "" {
		return fmt.Errorf("DB_NAME cannot be empty")
	}
	if c.DBUser == "" {
		return fmt.Errorf("DB_USER cannot be empty")
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

// GetDatabaseDSN returns the PostgreSQL database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
}