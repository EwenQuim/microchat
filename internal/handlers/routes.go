package handlers

import (
	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

func RegisterChatRoutes(s *fuego.Server, chatService *services.ChatService, cfg *config.Config) {

	// Room routes
	chatGroup := fuego.Group(s, "/rooms", option.TagInfo("chat", "routes relative to rooms and messaging"))
	fuego.Get(chatGroup, "", GetRooms(chatService))
	fuego.Get(chatGroup, "/search", SearchRooms(chatService))
	fuego.Post(chatGroup, "", CreateRoom(chatService))
	fuego.Get(chatGroup, "/{room}/messages", GetMessages(chatService))
	fuego.Post(chatGroup, "/{room}/messages", SendMessage(chatService))

	// User routes
	userGroup := fuego.Group(s, "/users", option.TagInfo("user", "routes relative to users"))
	fuego.Get(userGroup, "/{publicKey}", GetUser(chatService))
}
