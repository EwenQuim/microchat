package repository

import (
	"cmp"
	"fmt"
	"log/slog"
	"os"

	"github.com/EwenQuim/microchat/internal/repository/memory"
	"github.com/EwenQuim/microchat/internal/repository/sqlite"
	"github.com/EwenQuim/microchat/internal/services"
)

// NewRepository creates a new repository based on the DB_PATH environment variable.
// If DB_PATH is ":memory:" or empty, it uses in-memory storage.
// Otherwise, it uses SQLite with the specified path.
func NewRepository() (services.Repository, error) {
	dbPath := cmp.Or(os.Getenv("DB_PATH"), ":memory:")

	if dbPath == ":memory:" {
		slog.Info("Using in-memory storage. Set DB_PATH if you want persistent data")
		return memory.NewStore(), nil
	}

	slog.Info("Using SQLite database", "dbPath", dbPath)
	db, err := sqlite.InitDB(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SQLite database: %w", err)
	}

	return sqlite.NewStore(db), nil
}
