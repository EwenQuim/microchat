package handlers

import (
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func GetMessages(chatService *services.ChatService) func(c fuego.ContextNoBody) ([]models.Message, error) {
	return func(c fuego.ContextNoBody) ([]models.Message, error) {
		room := c.PathParam("room")
		return chatService.GetMessages(room)
	}
}

func SendMessage(chatService *services.ChatService) func(c fuego.ContextWithBody[models.SendMessageRequest]) (*models.Message, error) {
	return func(c fuego.ContextWithBody[models.SendMessageRequest]) (*models.Message, error) {
		room := c.PathParam("room")
		body, err := c.Body()
		if err != nil {
			return nil, err
		}

		return chatService.SendMessage(room, body.User, body.Content)
	}
}
