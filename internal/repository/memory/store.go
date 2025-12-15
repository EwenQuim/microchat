package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/EwenQuim/microchat/internal/models"
	"github.com/google/uuid"
)

type Store struct {
	mu       sync.RWMutex
	messages map[string][]models.Message
}

func NewStore() *Store {
	return &Store{
		messages: make(map[string][]models.Message),
	}
}

func (s *Store) SaveMessage(room, user, content, signature, pubkey string, signedTimestamp int64) (*models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *Store) GetMessages(room string) ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	messages, exists := s.messages[room]
	if !exists {
		return []models.Message{}, nil
	}

	return messages, nil
}

func (s *Store) GetRooms() ([]models.Room, error) {
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

func (s *Store) CreateRoom(name string) (*models.Room, error) {
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
