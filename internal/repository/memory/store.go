package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/google/uuid"
)

type roomMetadata struct {
	Hidden    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store struct {
	mu       sync.RWMutex
	messages map[string][]models.Message
	users    map[string]*models.User // username -> User
	rooms    map[string]*roomMetadata
}

// Ensure Store implements the Repository interface
var _ services.Repository = (*Store)(nil)

func NewStore() *Store {
	return &Store{
		messages: make(map[string][]models.Message),
		users:    make(map[string]*models.User),
		rooms:    make(map[string]*roomMetadata),
	}
}

func (s *Store) SaveMessage(ctx context.Context, room, user, content, signature, pubkey string, signedTimestamp int64) (*models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Automatically create room if it doesn't exist
	if _, exists := s.rooms[room]; !exists {
		now := time.Now()
		s.rooms[room] = &roomMetadata{
			Hidden:    false, // Default to visible
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.messages[room] = []models.Message{}
	}

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

	rooms := make([]models.Room, 0, len(s.rooms))
	for name, metadata := range s.rooms {
		messageCount := 0
		if messages, exists := s.messages[name]; exists {
			messageCount = len(messages)
		}
		rooms = append(rooms, models.Room{
			Name:         name,
			MessageCount: messageCount,
			Hidden:       metadata.Hidden,
		})
	}

	return rooms, nil
}

func (s *Store) SearchRooms(ctx context.Context, query string) ([]models.Room, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]models.Room, 0)
	for name, metadata := range s.rooms {
		// Simple case-insensitive substring search
		if query == "" || containsCaseInsensitive(name, query) {
			messageCount := 0
			var lastMessage *models.Message
			if messages, exists := s.messages[name]; exists {
				messageCount = len(messages)
				if messageCount > 0 {
					lastMessage = &messages[messageCount-1]
				}
			}

			room := models.Room{
				Name:         name,
				MessageCount: messageCount,
				Hidden:       metadata.Hidden,
			}

			if lastMessage != nil {
				room.LastMessageContent = &lastMessage.Content
				room.LastMessageUser = &lastMessage.User
				timestamp := lastMessage.Timestamp.Format(time.RFC3339)
				room.LastMessageTimestamp = &timestamp
			}

			rooms = append(rooms, room)
		}
	}

	return rooms, nil
}

func (s *Store) CreateRoom(ctx context.Context, name string) (*models.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[name]; exists {
		return nil, fmt.Errorf("room already exists")
	}

	now := time.Now()
	s.rooms[name] = &roomMetadata{
		Hidden:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.messages[name] = []models.Message{}
	return &models.Room{
		Name:         name,
		MessageCount: 0,
		Hidden:       false,
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

func (s *Store) GetUserWithPostCount(ctx context.Context, publicKey string) (*models.UserWithPostCount, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[publicKey]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Count posts by this user
	postCount := int64(0)
	for _, messages := range s.messages {
		for _, msg := range messages {
			if msg.Pubkey == publicKey {
				postCount++
			}
		}
	}

	return &models.UserWithPostCount{
		PublicKey: user.PublicKey,
		Verified:  user.Verified,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		PostCount: postCount,
	}, nil
}

func (s *Store) UpdateRoomVisibility(ctx context.Context, name string, hidden bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[name]
	if !exists {
		return fmt.Errorf("room not found")
	}

	room.Hidden = hidden
	room.UpdatedAt = time.Now()
	return nil
}

func containsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
