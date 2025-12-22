-- +goose Up
-- Remove the hidden field from rooms table
-- SQLite requires recreating the table to drop a column

-- Drop the index on hidden field
DROP INDEX IF EXISTS idx_rooms_hidden;

-- Create new table without hidden column
CREATE TABLE rooms_new (
    name TEXT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    password_hash TEXT
);

-- Copy data from old table (hidden field is dropped)
INSERT INTO rooms_new (name, created_at, updated_at, password_hash)
SELECT name, created_at, updated_at, password_hash FROM rooms;

-- Drop old table
DROP TABLE rooms;

-- Rename new table
ALTER TABLE rooms_new RENAME TO rooms;

-- Recreate the password index
CREATE INDEX IF NOT EXISTS idx_rooms_password ON rooms(password_hash) WHERE password_hash IS NOT NULL;

-- +goose Down
-- Recreate rooms table with hidden field (all rooms will be visible by default)

CREATE TABLE rooms_new (
    name TEXT PRIMARY KEY,
    hidden BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    password_hash TEXT
);

-- Copy data back (all rooms will be visible/non-hidden by default)
INSERT INTO rooms_new (name, hidden, created_at, updated_at, password_hash)
SELECT name, 0, created_at, updated_at, password_hash FROM rooms;

DROP TABLE rooms;
ALTER TABLE rooms_new RENAME TO rooms;

CREATE INDEX IF NOT EXISTS idx_rooms_hidden ON rooms(hidden);
CREATE INDEX IF NOT EXISTS idx_rooms_password ON rooms(password_hash) WHERE password_hash IS NOT NULL;
