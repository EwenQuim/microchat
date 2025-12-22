-- name: CreateMessage :one
INSERT INTO messages (id, room, user, content, timestamp, signature, pubkey, signed_timestamp)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetMessagesByRoom :many
SELECT * FROM messages
WHERE room = ?
ORDER BY timestamp ASC
LIMIT 100;

-- name: GetRoomsWithMessageCount :many
SELECT
    r.name,
    CASE WHEN r.password_hash IS NOT NULL THEN 1 ELSE 0 END as has_password,
    COALESCE(last_msg.content, '') as last_message_content,
    COALESCE(last_msg.user, '') as last_message_user,
    CASE
        WHEN last_msg.timestamp IS NULL THEN CAST('1970-01-01 00:00:00' AS TEXT)
        ELSE CAST(last_msg.timestamp AS TEXT)
    END as last_message_timestamp
FROM rooms r
LEFT JOIN (
    SELECT m.*
    FROM messages m
    INNER JOIN (
        SELECT room, MAX(timestamp) as max_timestamp
        FROM messages
        GROUP BY room
    ) latest ON m.room = latest.room AND m.timestamp = latest.max_timestamp
) last_msg ON last_msg.room = r.name
ORDER BY last_msg.timestamp DESC, r.name ASC
LIMIT 100;

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
SELECT * FROM users
LIMIT 100;

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
INSERT INTO rooms (name, password_hash, created_at, updated_at)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetRoomByName :one
SELECT * FROM rooms
WHERE name = ?;

-- name: RoomExists :one
SELECT COUNT(*) > 0 as room_exists FROM rooms WHERE name = ?;

-- name: SearchRoomsByName :many
SELECT
    r.name,
    CASE WHEN r.password_hash IS NOT NULL THEN 1 ELSE 0 END as has_password,
    COALESCE(last_msg.content, '') as last_message_content,
    COALESCE(last_msg.user, '') as last_message_user,
    CASE
        WHEN last_msg.timestamp IS NULL THEN CAST('1970-01-01 00:00:00' AS TEXT)
        ELSE CAST(last_msg.timestamp AS TEXT)
    END as last_message_timestamp
FROM rooms r
LEFT JOIN (
    SELECT m.*
    FROM messages m
    INNER JOIN (
        SELECT room, MAX(timestamp) as max_timestamp
        FROM messages
        GROUP BY room
    ) latest ON m.room = latest.room AND m.timestamp = latest.max_timestamp
) last_msg ON last_msg.room = r.name
WHERE r.name LIKE '%' || ? || '%'
ORDER BY last_message_timestamp DESC, r.name ASC
LIMIT 100;

-- name: GetRoomPasswordHash :one
SELECT password_hash FROM rooms WHERE name = ?;
