package sqlite

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// InitDB initializes a SQLite database connection and runs migrations
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations runs all pending migrations
func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close(db *sql.DB) error {
	if err := db.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}

// MigrateDown rolls back the last migration (useful for development)
func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Down(db, "migrations"); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// MigrationStatus returns the current migration status
func MigrationStatus(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Status(db, "migrations"); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}
