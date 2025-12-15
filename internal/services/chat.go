package services

import (
	"context"

	"github.com/EwenQuim/microchat/internal/models"
)

type Repository interface {
	SaveMessage(ctx context.Context, room, user, content, signature, pubkey string, timestamp int64) (*models.Message, error)
	GetMessages(ctx context.Context, room string) ([]models.Message, error)
	GetRooms(ctx context.Context) ([]models.Room, error)
	CreateRoom(ctx context.Context, name string) (*models.Room, error)

	// User management
	RegisterUser(ctx context.Context, publicKey string) (*models.User, error)
	GetUser(ctx context.Context, publicKey string) (*models.User, error)
	GetUserByPublicKey(ctx context.Context, publicKey string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]models.User, error)
	VerifyUser(ctx context.Context, publicKey string) error
	UnverifyUser(ctx context.Context, publicKey string) error
}

type ChatService struct {
	repo Repository
}

func NewChatService(repo Repository) *ChatService {
	return &ChatService{
		repo: repo,
	}
}

func (s *ChatService) SendMessage(ctx context.Context, room, user, content, signature, pubkey string, timestamp int64) (*models.Message, error) {
	return s.repo.SaveMessage(ctx, room, user, content, signature, pubkey, timestamp)
}

func (s *ChatService) GetMessages(ctx context.Context, room string) ([]models.Message, error) {
	return s.repo.GetMessages(ctx, room)
}

func (s *ChatService) GetRooms(ctx context.Context) ([]models.Room, error) {
	return s.repo.GetRooms(ctx)
}

func (s *ChatService) CreateRoom(ctx context.Context, name string) (*models.Room, error) {
	return s.repo.CreateRoom(ctx, name)
}

func (s *ChatService) RegisterUser(ctx context.Context, publicKey string) (*models.User, error) {
	return s.repo.RegisterUser(ctx, publicKey)
}

func (s *ChatService) GetUser(ctx context.Context, publicKey string) (*models.User, error) {
	return s.repo.GetUser(ctx, publicKey)
}

func (s *ChatService) GetUserByPublicKey(ctx context.Context, publicKey string) (*models.User, error) {
	return s.repo.GetUserByPublicKey(ctx, publicKey)
}

func (s *ChatService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.GetAllUsers(ctx)
}

func (s *ChatService) VerifyUser(ctx context.Context, publicKey string) error {
	return s.repo.VerifyUser(ctx, publicKey)
}

func (s *ChatService) UnverifyUser(ctx context.Context, publicKey string) error {
	return s.repo.UnverifyUser(ctx, publicKey)
}
