package services

import "github.com/EwenQuim/microchat/internal/models"

type Repository interface {
	SaveMessage(room, user, content, signature, pubkey string, timestamp int64) (*models.Message, error)
	GetMessages(room string) ([]models.Message, error)
	GetRooms() ([]models.Room, error)
	CreateRoom(name string) (*models.Room, error)

	// User management
	RegisterUser(publicKey string) (*models.User, error)
	GetUser(publicKey string) (*models.User, error)
	GetUserByPublicKey(publicKey string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	VerifyUser(publicKey string) error
	UnverifyUser(publicKey string) error
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

func (s *ChatService) RegisterUser(publicKey string) (*models.User, error) {
	return s.repo.RegisterUser(publicKey)
}

func (s *ChatService) GetUser(publicKey string) (*models.User, error) {
	return s.repo.GetUser(publicKey)
}

func (s *ChatService) GetUserByPublicKey(publicKey string) (*models.User, error) {
	return s.repo.GetUserByPublicKey(publicKey)
}

func (s *ChatService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAllUsers()
}

func (s *ChatService) VerifyUser(publicKey string) error {
	return s.repo.VerifyUser(publicKey)
}

func (s *ChatService) UnverifyUser(publicKey string) error {
	return s.repo.UnverifyUser(publicKey)
}
