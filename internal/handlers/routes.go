package handlers

import (
	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

func RegisterChatRoutes(s *fuego.Server, chatService *services.ChatService, cfg *config.Config) {

	// Room routes
	chatGroup := fuego.Group(s, "/rooms", option.Tags("chat"))
	fuego.Get(chatGroup, "", GetRooms(chatService))
	fuego.Post(chatGroup, "", CreateRoom(chatService))
	fuego.Patch(chatGroup, "/{room}/visibility", UpdateRoomVisibility(chatService, cfg))
	fuego.Get(chatGroup, "/{room}/messages", GetMessages(chatService))
	fuego.Post(chatGroup, "/{room}/messages", SendMessage(chatService))

	// User routes

	userGroup := fuego.Group(s, "/users")
	fuego.Post(userGroup, "", RegisterUser(chatService))
	fuego.Get(userGroup, "", GetAllUsers(chatService))
	fuego.Get(userGroup, "/{publicKey}", GetUser(chatService))
	fuego.Get(userGroup, "/{publicKey}/details", GetUserDetails(chatService))
	fuego.Post(userGroup, "/verify", VerifyUser(chatService))
	fuego.Post(userGroup, "/unverify", UnverifyUser(chatService))
}
