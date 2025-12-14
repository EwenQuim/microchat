package handlers

import (
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func GetRooms(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.Room, error) {
	return func(c fuego.ContextNoBody) ([]models.Room, error) {
		return chatService.GetRooms()
	}
}
