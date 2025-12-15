package handlers

import (
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func RegisterChatRoutes(s *fuego.Server, chatService *services.ChatService) {
	fuego.Get(s, "/rooms", GetRooms(chatService))
	fuego.Post(s, "/rooms", CreateRoom(chatService))
	fuego.Get(s, "/rooms/{room}/messages", GetMessages(chatService))
	fuego.Post(s, "/rooms/{room}/messages", SendMessage(chatService))
}
