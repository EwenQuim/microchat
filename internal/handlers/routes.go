package handlers

import (
	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/services"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/jub0bs/cors"
)

func RegisterChatRoutes(s *fuego.Server, chatService *services.ChatService, cfg *config.Config) {
	corsMw, err := cors.NewMiddleware(cors.Config{
		Origins:        []string{"*"},
		Methods:        []string{"GET", "POST"},
		RequestHeaders: []string{"Content-Type"},
	})
	if err != nil {
		panic(err)
	}
	fuego.Use(s, corsMw.Wrap)

	// Server info
	fuego.Get(s, "/server-info", GetServerInfo(cfg))

	// Room routes
	chatGroup := fuego.Group(s, "/rooms", option.TagInfo("chat", "routes relative to rooms and messaging"))
	fuego.Get(chatGroup, "", GetRooms(chatService))
	fuego.Get(chatGroup, "/search", SearchRooms(chatService))
	fuego.Post(chatGroup, "", CreateRoom(chatService),
		option.RequestContentType("application/json"),
	)
	fuego.Get(chatGroup, "/{room}/messages", GetMessages(chatService))
	fuego.Post(chatGroup, "/{room}/messages", SendMessage(chatService),
		option.RequestContentType("application/json"),
	)

	// User routes
	userGroup := fuego.Group(s, "/users", option.TagInfo("user", "routes relative to users"))
	fuego.Get(userGroup, "/{publicKey}", GetUser(chatService))
}
