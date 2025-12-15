-- +goose Up
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    public_key TEXT PRIMARY KEY,
    verified BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_users_verified ON users(verified);

-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    room TEXT NOT NULL,
    user TEXT NOT NULL,
    content TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    signature TEXT,
    pubkey TEXT,
    signed_timestamp INTEGER,
    FOREIGN KEY (pubkey) REFERENCES users(public_key) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_messages_pubkey ON messages(pubkey);

-- +goose Down
DROP INDEX IF EXISTS idx_messages_pubkey;
DROP INDEX IF EXISTS idx_messages_timestamp;
DROP INDEX IF EXISTS idx_messages_room;
DROP TABLE IF EXISTS messages;

DROP INDEX IF EXISTS idx_users_verified;
DROP TABLE IF EXISTS users;
