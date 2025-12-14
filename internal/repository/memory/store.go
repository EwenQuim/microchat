package memory

import (
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

func (s *Store) SaveMessage(room, user, content string) (*models.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := models.Message{
		ID:        uuid.New().String(),
		Room:      room,
		User:      user,
		Content:   content,
		Timestamp: time.Now(),
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
