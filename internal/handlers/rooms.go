package handlers

import (
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func GetRooms(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.Room, error) {
	return func(c fuego.ContextNoBody) ([]models.Room, error) {
		return chatService.GetRooms(c.Context())
	}
}

func CreateRoom(chatService *services.ChatService) func(c fuego.ContextWithBody[models.CreateRoomRequest]) (*models.Room, error) {
	return func(c fuego.ContextWithBody[models.CreateRoomRequest]) (*models.Room, error) {
		body, err := c.Body()
		if err != nil {
			return nil, err
		}
		return chatService.CreateRoom(c.Context(), body.Name)
	}
}
