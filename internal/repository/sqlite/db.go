package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB initializes a SQLite database connection and creates the schema
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create schema
	if err := createSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}

// createSchema creates the necessary tables for the application
func createSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		room TEXT NOT NULL,
		user TEXT NOT NULL,
		content TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		signature TEXT,
		pubkey TEXT,
		signed_timestamp INTEGER,
		FOREIGN KEY (pubkey) REFERENCES users(public_key)
	);

	CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room);
	CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
	CREATE INDEX IF NOT EXISTS idx_messages_pubkey ON messages(pubkey);

	CREATE TABLE IF NOT EXISTS users (
		public_key TEXT PRIMARY KEY,
		verified BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_users_verified ON users(verified);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
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
