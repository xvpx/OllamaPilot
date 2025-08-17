package database

import (
	"fmt"

	"chat_ollama/internal/config"
)

// NewDatabase creates a new PostgreSQL database connection
func NewDatabase(cfg *config.Config) (Database, error) {
	return NewPostgresDB(cfg)
}

// RunMigrations runs PostgreSQL migrations
func RunMigrations(db Database, cfg *config.Config) error {
	if pgDB, ok := db.(*PostgresDB); ok {
		return pgDB.RunMigrations("migrations/postgres")
	}
	return fmt.Errorf("invalid PostgreSQL database instance")
}