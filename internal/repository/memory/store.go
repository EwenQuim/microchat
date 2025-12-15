package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/google/uuid"
)

type Store struct {
	mu       sync.RWMutex
	messages map[string][]models.Message
	users    map[string]*models.User // username -> User
}

// Ensure Store implements the Repository interface
var _ services.Repository = (*Store)(nil)

func NewStore() *Store {
	return &Store{
		messages: make(map[string][]models.Message),
		users:    make(map[string]*models.User),
	}
}

func (s *Store) SaveMessage(ctx context.Context, room, user, content, signature, pubkey string, signedTimestamp int64) (*models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Automatically create unverified user if pubkey is provided and user doesn't exist
	if pubkey != "" {
		if _, exists := s.users[pubkey]; !exists {
			now := time.Now()
			s.users[pubkey] = &models.User{
				PublicKey: pubkey,
				Verified:  false,
				CreatedAt: now,
				UpdatedAt: now,
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

	s.messages[room] = append(s.messages[room], msg)
	return &msg, nil
}

func (s *Store) GetMessages(ctx context.Context, room string) ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages, exists := s.messages[room]
	if !exists {
		return []models.Message{}, nil
	}

	return messages, nil
}

func (s *Store) GetRooms(ctx context.Context) ([]models.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]models.Room, 0, len(s.messages))
	for name, messages := range s.messages {
		rooms = append(rooms, models.Room{
			Name:         name,
			MessageCount: len(messages),
		})
	}

	return rooms, nil
}

func (s *Store) CreateRoom(ctx context.Context, name string) (*models.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.messages[name]; exists {
		return nil, fmt.Errorf("room already exists")
	}

	s.messages[name] = []models.Message{}
	return &models.Room{
		Name:         name,
		MessageCount: 0,
	}, nil
}

func (s *Store) RegisterUser(ctx context.Context, publicKey string) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if public key is already registered
	for _, user := range s.users {
		if user.PublicKey == publicKey {
			return nil, fmt.Errorf("public key already registered to user %s", publicKey)
		}
	}

	now := time.Now()
	user := &models.User{
		PublicKey: publicKey,
		Verified:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return user, nil
}

func (s *Store) GetUser(ctx context.Context, publicKey string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[publicKey]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (s *Store) GetUserByPublicKey(ctx context.Context, publicKey string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.PublicKey == publicKey {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}

func (s *Store) GetAllUsers(ctx context.Context) ([]models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, *user)
	}

	return users, nil
}

func (s *Store) VerifyUser(ctx context.Context, publicKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[publicKey]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.Verified = true
	user.UpdatedAt = time.Now()
	return nil
}

func (s *Store) UnverifyUser(ctx context.Context, publicKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[publicKey]
	if !exists {
		return fmt.Errorf("user not found")
	}

	user.Verified = false
	user.UpdatedAt = time.Now()
	return nil
}
