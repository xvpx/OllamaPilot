package database

import (
	"context"
	"database/sql"
)

// Database interface defines the common database operations
type Database interface {
	// Basic database operations
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error

	// Health and version operations
	HealthCheck() error
	GetVersion() (string, error)

	// Migration operations
	RunMigrations(migrationsPath string) error

	// Vector operations (PostgreSQL specific)
	VectorSimilaritySearch(table, vectorColumn string, queryVector []float32, limit int) (*sql.Rows, error)
	InsertVector(table string, columns []string, values []interface{}) error
	CheckPgvectorExtension() error
}