package database

import (
	"fmt"

	"chat_ollama/internal/config"
)

// NewDatabase creates a new database connection based on configuration
func NewDatabase(cfg *config.Config) (Database, error) {
	switch cfg.DBType {
	case "postgres":
		return NewPostgresDB(cfg)
	case "sqlite":
		return NewSQLiteDB(cfg.DBPath)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.DBType)
	}
}

// RunMigrations runs the appropriate migrations based on database type
func RunMigrations(db Database, cfg *config.Config) error {
	switch cfg.DBType {
	case "postgres":
		if pgDB, ok := db.(*PostgresDB); ok {
			return pgDB.RunMigrations("migrations/postgres")
		}
		return fmt.Errorf("invalid PostgreSQL database instance")
	case "sqlite":
		if sqliteDB, ok := db.(*DB); ok {
			return sqliteDB.RunMigrations()
		}
		return fmt.Errorf("invalid SQLite database instance")
	default:
		return fmt.Errorf("unsupported database type for migrations: %s", cfg.DBType)
	}
}