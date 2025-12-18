-- +goose Up
-- Create rooms table
CREATE TABLE IF NOT EXISTS rooms (
    name TEXT PRIMARY KEY,
    hidden BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rooms_hidden ON rooms(hidden);

-- +goose Down
DROP INDEX IF EXISTS idx_rooms_hidden;
DROP TABLE IF EXISTS rooms;
