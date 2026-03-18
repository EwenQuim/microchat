package handlers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/go-fuego/fuego"
)

// stubRepo is a no-op repository for testing.
type stubRepo struct{}

func (s *stubRepo) SaveMessage(_ context.Context, _, _, _, _, _ string, _ int64) (*models.Message, error) {
	return &models.Message{}, nil
}
func (s *stubRepo) GetMessages(_ context.Context, _ string, _ services.MessageQueryParams) ([]models.Message, error) {
	return nil, nil
}
func (s *stubRepo) GetRooms(_ context.Context) ([]models.Room, error) { return nil, nil }
func (s *stubRepo) SearchRooms(_ context.Context, _ string) ([]models.Room, error) {
	return nil, nil
}
func (s *stubRepo) CreateRoom(_ context.Context, _ string, _ *string) (*models.Room, error) {
	return nil, nil
}
func (s *stubRepo) ValidateRoomPassword(_ context.Context, _, _ string) error { return nil }
func (s *stubRepo) RegisterUser(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}
func (s *stubRepo) GetUser(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}
func (s *stubRepo) GetUserByPublicKey(_ context.Context, _ string) (*models.User, error) {
	return nil, nil
}
func (s *stubRepo) GetUserWithPostCount(_ context.Context, _ string) (*models.UserWithPostCount, error) {
	return nil, nil
}
func (s *stubRepo) GetAllUsers(_ context.Context) ([]models.User, error) { return nil, nil }
func (s *stubRepo) VerifyUser(_ context.Context, _ string) error         { return nil }
func (s *stubRepo) UnverifyUser(_ context.Context, _ string) error       { return nil }

func newTestServer(t *testing.T) *fuego.Server {
	t.Helper()
	chatService := services.NewChatService(&stubRepo{})
	s := fuego.NewServer(fuego.WithoutLogger())
	apiGroup := fuego.Group(s, "/api")
	RegisterChatRoutes(apiGroup, chatService, &config.Config{})
	return s
}

func TestGetMessages_InvalidBefore_Returns400(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/rooms/test/messages?before=not-a-date", nil)
	w := httptest.NewRecorder()

	s.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for invalid before param; body: %s", w.Code, w.Body.String())
	}
}

func TestGetMessages_ValidBefore_Returns200(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/rooms/test/messages?before=2025-01-01T00:00:00Z&limit=10", nil)
	w := httptest.NewRecorder()

	s.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 for valid before param; body: %s", w.Code, w.Body.String())
	}
}

func TestSendMessage_MissingSignature_Returns400(t *testing.T) {
	s := newTestServer(t)

	body := []byte(`{"user":"alice","content":"hello"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/test/messages", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	s.Mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400; body: %s", w.Code, w.Body.String())
	}
}
