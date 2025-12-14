package handlers

import (
	"github.com/EwenQuim/microchat/internal/models"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/go-fuego/fuego"
)

func RegisterChatRoutes(api *fuego.Server, chatService *services.ChatService) {
	fuego.Get(api, "/rooms", func(c fuego.ContextNoBody) ([]models.Room, error) {
		return chatService.GetRooms()
	})

	fuego.Get(api, "/rooms/{room}/messages", func(c fuego.ContextNoBody) ([]models.Message, error) {
		room := c.PathParam("room")
		return chatService.GetMessages(room)
	})

	fuego.Post(api, "/rooms/{room}/messages", func(c fuego.ContextWithBody[models.SendMessageRequest]) (*models.Message, error) {
		room := c.PathParam("room")
		body, err := c.Body()
		if err != nil {
			return nil, err
		}

		return chatService.SendMessage(room, body.User, body.Content)
	})
}
