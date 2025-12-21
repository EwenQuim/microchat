-- +goose Up
-- +goose StatementBegin
ALTER TABLE rooms ADD COLUMN password_hash TEXT;
CREATE INDEX IF NOT EXISTS idx_rooms_password ON rooms(password_hash) WHERE password_hash IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_rooms_password;
ALTER TABLE rooms DROP COLUMN password_hash;
-- +goose StatementEnd
