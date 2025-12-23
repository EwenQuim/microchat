-- +goose Up
-- Add composite index for efficient last message queries
CREATE INDEX IF NOT EXISTS idx_messages_room_timestamp ON messages(room, timestamp DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_messages_room_timestamp;
