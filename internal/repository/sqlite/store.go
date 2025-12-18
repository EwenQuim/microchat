package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/repository/sqlite/sqlc"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/google/uuid"
)

type Store struct {
	queries *sqlc.Queries
}

// Ensure Store implements the Repository interface
var _ services.Repository = (*Store)(nil)

func NewStore(db *sql.DB) *Store {
	return &Store{
		queries: sqlc.New(db),
	}
}

func (s *Store) SaveMessage(ctx context.Context, room, user, content, signature, pubkey string, signedTimestamp int64) (*models.Message, error) {

	// Automatically create unverified user if pubkey is provided and user doesn't exist
	if pubkey != "" {
		exists, err := s.queries.UserExistsByPublicKey(ctx, pubkey)
		if err != nil {
			return nil, fmt.Errorf("failed to check user existence: %w", err)
		}

		if !exists {
			now := time.Now()
			_, err = s.queries.CreateUser(ctx, sqlc.CreateUserParams{
				PublicKey: pubkey,
				Verified:  false,
				CreatedAt: now,
				UpdatedAt: now,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create user: %w", err)
			}
		}
	}

	msgID := uuid.New().String()
	timestamp := time.Now()

	sqlcMsg, err := s.queries.CreateMessage(ctx, sqlc.CreateMessageParams{
		ID:        msgID,
		Room:      room,
		User:      user,
		Content:   content,
		Timestamp: timestamp,
		Signature: sql.NullString{
			String: signature,
			Valid:  signature != "",
		},
		Pubkey: sql.NullString{
			String: pubkey,
			Valid:  pubkey != "",
		},
		SignedTimestamp: sql.NullInt64{
			Int64: signedTimestamp,
			Valid: signedTimestamp != 0,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	return sqlcMessageToModel(sqlcMsg), nil
}

func (s *Store) GetMessages(ctx context.Context, room string) ([]models.Message, error) {

	sqlcMessages, err := s.queries.GetMessagesByRoom(ctx, room)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	messages := make([]models.Message, 0, len(sqlcMessages))
	for _, msg := range sqlcMessages {
		messages = append(messages, *sqlcMessageToModel(msg))
	}

	return messages, nil
}

func (s *Store) GetRooms(ctx context.Context) ([]models.Room, error) {
	rows, err := s.queries.GetRoomsWithMessageCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms: %w", err)
	}

	rooms := make([]models.Room, 0, len(rows))
	for _, row := range rows {
		rooms = append(rooms, models.Room{
			Name:         row.Room,
			MessageCount: int(row.MessageCount),
		})
	}

	return rooms, nil
}

func (s *Store) CreateRoom(ctx context.Context, name string) (*models.Room, error) {
	// Check if room already exists (has messages)
	count, err := s.queries.GetMessageCountByRoom(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to check room existence: %w", err)
	}

	if count > 0 {
		return nil, fmt.Errorf("room already exists")
	}

	return &models.Room{
		Name:         name,
		MessageCount: 0,
	}, nil
}

func (s *Store) RegisterUser(ctx context.Context, publicKey string) (*models.User, error) {

	// Check if public key is already registered
	exists, err := s.queries.UserExistsByPublicKey(ctx, publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("public key already registered to user %s", publicKey)
	}

	now := time.Now()
	sqlcUser, err := s.queries.CreateUser(ctx, sqlc.CreateUserParams{
		PublicKey: publicKey,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	return sqlcUserToModel(sqlcUser), nil
}

func (s *Store) GetUser(ctx context.Context, publicKey string) (*models.User, error) {

	sqlcUser, err := s.queries.GetUserByPublicKey(ctx, publicKey)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return sqlcUserToModel(sqlcUser), nil
}

func (s *Store) GetUserByPublicKey(ctx context.Context, publicKey string) (*models.User, error) {
	return s.GetUser(ctx, publicKey)
}

func (s *Store) GetAllUsers(ctx context.Context) ([]models.User, error) {

	sqlcUsers, err := s.queries.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	users := make([]models.User, 0, len(sqlcUsers))
	for _, user := range sqlcUsers {
		users = append(users, *sqlcUserToModel(user))
	}

	return users, nil
}

func (s *Store) VerifyUser(ctx context.Context, publicKey string) error {

	err := s.queries.UpdateUserVerified(ctx, sqlc.UpdateUserVerifiedParams{
		Verified:  true,
		UpdatedAt: time.Now(),
		PublicKey: publicKey,
	})
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}

	return nil
}

func (s *Store) UnverifyUser(ctx context.Context, publicKey string) error {

	err := s.queries.UpdateUserVerified(ctx, sqlc.UpdateUserVerifiedParams{
		Verified:  false,
		UpdatedAt: time.Now(),
		PublicKey: publicKey,
	})
	if err != nil {
		return fmt.Errorf("failed to unverify user: %w", err)
	}

	return nil
}

// Helper functions to convert between sqlc and models types
func sqlcMessageToModel(msg sqlc.Message) *models.Message {
	return &models.Message{
		ID:              msg.ID,
		Room:            msg.Room,
		User:            msg.User,
		Content:         msg.Content,
		Timestamp:       msg.Timestamp,
		Signature:       msg.Signature.String,
		Pubkey:          msg.Pubkey.String,
		SignedTimestamp: msg.SignedTimestamp.Int64,
	}
}

func sqlcUserToModel(user sqlc.User) *models.User {
	return &models.User{
		PublicKey: user.PublicKey,
		Verified:  user.Verified,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func (s *Store) GetUserWithPostCount(ctx context.Context, publicKey string) (*models.UserWithPostCount, error) {
	row, err := s.queries.GetUserWithPostCount(ctx, publicKey)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user with post count: %w", err)
	}

	postCount := int64(0)
	if row.PostCount != nil {
		if count, ok := row.PostCount.(int64); ok {
			postCount = count
		}
	}

	return &models.UserWithPostCount{
		PublicKey: row.PublicKey,
		Verified:  row.Verified,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		PostCount: postCount,
	}, nil
}
