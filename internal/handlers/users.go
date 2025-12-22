package handlers

import (
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func GetUser(chatService *services.ChatService) func(c fuego.ContextNoBody) (*models.User, error) {
	return func(c fuego.ContextNoBody) (*models.User, error) {
		publicKey := c.PathParam("publicKey")
		return chatService.GetUser(c.Context(), publicKey)
	}
}

