package handlers

import (
	"net/http"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RegisterChatRoutes(s *fuego.Server, chatService *services.ChatService, cfg *config.Config) {
	fuego.Use(s, corsMiddleware)

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
