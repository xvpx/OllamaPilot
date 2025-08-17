package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"chat_ollama/internal/config"

	_ "github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

// PostgresDB wraps the sql.DB connection for PostgreSQL
type PostgresDB struct {
	*sql.DB
	config *config.Config
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg *config.Config) (*PostgresDB, error) {
	dsn := cfg.GetDatabaseDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL database: %w", err)
	}

	// Configure connection pool for PostgreSQL
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // No limit

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return &PostgresDB{
		DB:     db,
		config: cfg,
	}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

// HealthCheck performs a health check on the database
func (db *PostgresDB) HealthCheck() error {
	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("PostgreSQL health check failed: %w", err)
	}
	if result != 1 {
		return fmt.Errorf("PostgreSQL health check returned unexpected result: %d", result)
	}
	return nil
}

// GetVersion returns the PostgreSQL version
func (db *PostgresDB) GetVersion() (string, error) {
	var version string
	err := db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("failed to get PostgreSQL version: %w", err)
	}
	return version, nil
}

// CheckPgvectorExtension checks if pgvector extension is available
func (db *PostgresDB) CheckPgvectorExtension() error {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check pgvector extension: %w", err)
	}
	if !exists {
		return fmt.Errorf("pgvector extension is not installed")
	}
	return nil
}

// RunMigrations runs PostgreSQL migrations from the migrations/postgres directory
func (db *PostgresDB) RunMigrations(migrationsPath string) error {
	// Create migrations table if it doesn't exist
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationFiles, err := getMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Apply migrations
	for _, file := range migrationFiles {
		version := strings.TrimSuffix(filepath.Base(file), ".sql")
		
		// Check if migration already applied
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}
		
		if exists {
			continue // Skip already applied migration
		}

		// Read and execute migration
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Execute migration in a transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", version, err)
		}

		_, err = tx.Exec(string(content))
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		// Record migration as applied
		_, err = tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	return nil
}

// getMigrationFiles returns a sorted list of migration files
func getMigrationFiles(migrationsPath string) ([]string, error) {
	var files []string
	
	err := filepath.Walk(migrationsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".sql") {
			files = append(files, path)
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	sort.Strings(files)
	return files, nil
}

// VectorSimilaritySearch performs vector similarity search
func (db *PostgresDB) VectorSimilaritySearch(table, vectorColumn string, queryVector []float32, limit int) (*sql.Rows, error) {
	query := fmt.Sprintf(`
		SELECT *, (%s <=> $1) as distance 
		FROM %s 
		ORDER BY %s <=> $1 
		LIMIT $2
	`, vectorColumn, table, vectorColumn)
	
	return db.Query(query, pgvector.NewVector(queryVector), limit)
}

// InsertVector inserts a vector into the specified table
func (db *PostgresDB) InsertVector(table string, columns []string, values []interface{}) error {
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)
	
	_, err := db.Exec(query, values...)
	return err
}