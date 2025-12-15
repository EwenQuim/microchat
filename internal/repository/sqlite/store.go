package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

// Ensure Store implements the Repository interface
var _ services.Repository = (*Store)(nil)

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) SaveMessage(room, user, content, signature, pubkey string, signedTimestamp int64) (*models.Message, error) {
	// Automatically create unverified user if pubkey is provided and user doesn't exist
	if pubkey != "" {
		var exists bool
		err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE public_key = ?)", pubkey).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("failed to check user existence: %w", err)
		}

		if !exists {
			now := time.Now()
			_, err = s.db.Exec(
				"INSERT INTO users (public_key, verified, created_at, updated_at) VALUES (?, ?, ?, ?)",
				pubkey, false, now, now,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		}
	}

	msg := models.Message{
		ID:              uuid.New().String(),
		Room:            room,
		User:            user,
		Content:         content,
		Timestamp:       time.Now(),
		Signature:       signature,
		Pubkey:          pubkey,
		SignedTimestamp: signedTimestamp,
	}

	_, err := s.db.Exec(
		"INSERT INTO messages (id, room, user, content, timestamp, signature, pubkey, signed_timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		msg.ID, msg.Room, msg.User, msg.Content, msg.Timestamp, msg.Signature, msg.Pubkey, msg.SignedTimestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return &msg, nil
}

func (s *Store) GetMessages(room string) ([]models.Message, error) {
	rows, err := s.db.Query(
		"SELECT id, room, user, content, timestamp, signature, pubkey, signed_timestamp FROM messages WHERE room = ? ORDER BY timestamp ASC",
		room,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.Room, &msg.User, &msg.Content, &msg.Timestamp, &msg.Signature, &msg.Pubkey, &msg.SignedTimestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	if messages == nil {
		return []models.Message{}, nil
	}

	return messages, nil
}

func (s *Store) GetRooms() ([]models.Room, error) {
	rows, err := s.db.Query(
		"SELECT room, COUNT(*) as count FROM messages GROUP BY room",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query rooms: %w", err)
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
		err := rows.Scan(&room.Name, &room.MessageCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}
		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rooms: %w", err)
	}

	if rooms == nil {
		return []models.Room{}, nil
	}

	return rooms, nil
}

func (s *Store) CreateRoom(name string) (*models.Room, error) {
	// Check if room already exists (has messages)
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM messages WHERE room = ?", name).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to check room existence: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("room already exists")
	}

	// SQLite doesn't require explicit room creation, but we can insert a marker if needed
	// For now, we'll just return the room object
	return &models.Room{
		Name:         name,
		MessageCount: 0,
	}, nil
}

func (s *Store) RegisterUser(publicKey string) (*models.User, error) {
	// Check if public key is already registered
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE public_key = ?)", publicKey).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("public key already registered to user %s", publicKey)
	}

	now := time.Now()
	_, err = s.db.Exec(
		"INSERT INTO users (public_key, verified, created_at, updated_at) VALUES (?, ?, ?, ?)",
		publicKey, false, now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	user := &models.User{
		PublicKey: publicKey,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return user, nil
}

func (s *Store) GetUser(publicKey string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT public_key, verified, created_at, updated_at FROM users WHERE public_key = ?",
		publicKey,
	).Scan(&user.PublicKey, &user.Verified, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *Store) GetUserByPublicKey(publicKey string) (*models.User, error) {
	// Same as GetUser since public_key is the identifier
	return s.GetUser(publicKey)
}

func (s *Store) GetAllUsers() ([]models.User, error) {
	rows, err := s.db.Query(
		"SELECT public_key, verified, created_at, updated_at FROM users",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.PublicKey, &user.Verified, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	if users == nil {
		return []models.User{}, nil
	}

	return users, nil
}

func (s *Store) VerifyUser(publicKey string) error {
	result, err := s.db.Exec(
		"UPDATE users SET verified = ?, updated_at = ? WHERE public_key = ?",
		true, time.Now(), publicKey,
	)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (s *Store) UnverifyUser(publicKey string) error {
	result, err := s.db.Exec(
		"UPDATE users SET verified = ?, updated_at = ? WHERE public_key = ?",
		false, time.Now(), publicKey,
	)
	if err != nil {
		return fmt.Errorf("failed to unverify user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
