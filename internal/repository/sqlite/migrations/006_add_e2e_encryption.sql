-- +goose Up
-- +goose StatementBegin
ALTER TABLE rooms ADD COLUMN is_encrypted BOOLEAN DEFAULT FALSE;
ALTER TABLE rooms ADD COLUMN encryption_salt TEXT;

ALTER TABLE messages ADD COLUMN is_encrypted BOOLEAN DEFAULT FALSE;
ALTER TABLE messages ADD COLUMN nonce TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages DROP COLUMN nonce;
ALTER TABLE messages DROP COLUMN is_encrypted;

ALTER TABLE rooms DROP COLUMN encryption_salt;
ALTER TABLE rooms DROP COLUMN is_encrypted;
-- +goose StatementEnd
