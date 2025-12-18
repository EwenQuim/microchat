-- name: CreateMessage :one
INSERT INTO messages (id, room, user, content, timestamp, signature, pubkey, signed_timestamp)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMessagesByRoom :many
SELECT * FROM messages
WHERE room = ?
ORDER BY timestamp ASC;

-- name: GetRoomsWithMessageCount :many
SELECT
    r.name,
    r.hidden,
    COALESCE((SELECT COUNT(*) FROM messages WHERE room = r.name), 0) as message_count
FROM rooms r
ORDER BY r.name;

-- name: GetMessageCountByRoom :one
SELECT COUNT(*) as count
FROM messages
WHERE room = ?;

-- name: CreateUser :one
INSERT INTO users (public_key, verified, created_at, updated_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetUserByPublicKey :one
SELECT * FROM users
WHERE public_key = ?;

-- name: UserExistsByPublicKey :one
SELECT COUNT(*) > 0 as user_exists FROM users WHERE public_key = ?;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: UpdateUserVerified :exec
UPDATE users
SET verified = ?, updated_at = ?
WHERE public_key = ?;

-- name: GetUserVerified :one
SELECT verified FROM users
WHERE public_key = ?;

-- name: GetUserWithPostCount :one
SELECT
    u.*,
    COALESCE((SELECT COUNT(*) FROM messages WHERE pubkey = u.public_key), 0) as post_count
FROM users u
WHERE u.public_key = ?;

-- name: CreateRoom :one
INSERT INTO rooms (name, hidden, created_at, updated_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetRoomByName :one
SELECT * FROM rooms
WHERE name = ?;

-- name: RoomExists :one
SELECT COUNT(*) > 0 as room_exists FROM rooms WHERE name = ?;

-- name: UpdateRoomVisibility :exec
UPDATE rooms
SET hidden = ?, updated_at = ?
WHERE name = ?;
