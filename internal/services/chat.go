package services

import "github.com/EwenQuim/microchat/internal/models"

type Repository interface {
	SaveMessage(room, user, content, signature, pubkey string, timestamp int64) (*models.Message, error)
	GetMessages(room string) ([]models.Message, error)
	GetRooms() ([]models.Room, error)
	CreateRoom(name string) (*models.Room, error)
}

type ChatService struct {
	repo Repository
}

func NewChatService(repo Repository) *ChatService {
	return &ChatService{
		repo: repo,
	}
}

func (s *ChatService) SendMessage(room, user, content, signature, pubkey string, timestamp int64) (*models.Message, error) {
	return s.repo.SaveMessage(room, user, content, signature, pubkey, timestamp)
}

func (s *ChatService) GetMessages(room string) ([]models.Message, error) {
	return s.repo.GetMessages(room)
}

func (s *ChatService) GetRooms() ([]models.Room, error) {
	return s.repo.GetRooms()
}

func (s *ChatService) CreateRoom(name string) (*models.Room, error) {
	return s.repo.CreateRoom(name)
}
