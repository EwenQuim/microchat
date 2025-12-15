package handlers

import (
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func RegisterChatRoutes(s *fuego.Server, chatService *services.ChatService) {
	// Room routes
	fuego.Get(s, "/rooms", GetRooms(chatService))
	fuego.Post(s, "/rooms", CreateRoom(chatService))
	fuego.Get(s, "/rooms/{room}/messages", GetMessages(chatService))
	fuego.Post(s, "/rooms/{room}/messages", SendMessage(chatService))

	// User routes
	fuego.Post(s, "/users", RegisterUser(chatService))
	fuego.Get(s, "/users", GetAllUsers(chatService))
	fuego.Get(s, "/users/{publicKey}", GetUser(chatService))
	fuego.Post(s, "/users/verify", VerifyUser(chatService))
	fuego.Post(s, "/users/unverify", UnverifyUser(chatService))
}
