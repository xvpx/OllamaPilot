package database

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db *PostgresDB
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *PostgresDB) *MigrationRunner {
	return &MigrationRunner{db: db}
}

// RunMigrations runs all pending migrations
func (mr *MigrationRunner) RunMigrations(migrationsPath string) error {
	// Create migrations table if it doesn't exist
	if err := mr.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := mr.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Load migration files
	migrations, err := mr.loadMigrationFiles(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to load migration files: %w", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		if !appliedMigrations[migration.Version] {
			if err := mr.applyMigration(migration); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func (mr *MigrationRunner) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := mr.db.Exec(query)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func (mr *MigrationRunner) getAppliedMigrations() (map[int]bool, error) {
	query := "SELECT version FROM migrations"
	rows, err := mr.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// loadMigrationFiles loads migration files from the filesystem
func (mr *MigrationRunner) loadMigrationFiles(migrationsPath string) ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir(migrationsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		migration, err := mr.parseMigrationFile(path)
		if err != nil {
			return fmt.Errorf("failed to parse migration file %s: %w", path, err)
		}

		migrations = append(migrations, migration)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFile parses a migration file and extracts version and SQL
func (mr *MigrationRunner) parseMigrationFile(path string) (Migration, error) {
	// Extract version from filename (e.g., "001_initial_schema.sql" -> 1)
	filename := filepath.Base(path)
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return Migration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	var version int
	if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
		return Migration{}, fmt.Errorf("failed to parse version from filename %s: %w", filename, err)
	}

	// Read SQL content
	content, err := os.ReadFile(path)
	if err != nil {
		return Migration{}, fmt.Errorf("failed to read migration file: %w", err)
	}

	name := strings.TrimSuffix(parts[1], ".sql")

	return Migration{
		Version: version,
		Name:    name,
		SQL:     string(content),
	}, nil
}

// applyMigration applies a single migration
func (mr *MigrationRunner) applyMigration(migration Migration) error {
	tx, err := mr.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration as applied
	query := "INSERT INTO migrations (version, name) VALUES ($1, $2)"
	if _, err := tx.Exec(query, migration.Version, migration.Name); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}