package handlers

import (
	"time"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/middleware"
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

	// Rate limiters: one per window duration.
	minuteRL := middleware.NewRateLimiter(time.Minute)
	hourRL := middleware.NewRateLimiter(time.Hour)

	// Server info
	fuego.Get(s, "/server-info", GetServerInfo(cfg))

	// Room routes
	chatGroup := fuego.Group(s, "/rooms", option.TagInfo("chat", "routes relative to rooms and messaging"))

	fuego.Get(chatGroup, "", GetRooms(chatService),
		option.Middleware(middleware.IPRateLimit(minuteRL, 120, time.Minute)),
	)
	fuego.Get(chatGroup, "/search", SearchRooms(chatService),
		option.Middleware(middleware.IPRateLimit(minuteRL, 120, time.Minute)),
	)
	fuego.Post(chatGroup, "", CreateRoom(chatService),
		option.RequestContentType("application/json"),
		option.Middleware(middleware.IPRateLimit(hourRL, 10, time.Hour)),
	)
	fuego.Get(chatGroup, "/{room}/messages", GetMessages(chatService, minuteRL),
		option.Middleware(middleware.IPRateLimit(minuteRL, 60, time.Minute)),
	)
	fuego.Post(chatGroup, "/{room}/messages", SendMessage(chatService, minuteRL),
		option.RequestContentType("application/json"),
		option.Middleware(middleware.MessageRateLimit(minuteRL, 20, 30, time.Minute)),
	)

	// User routes
	userGroup := fuego.Group(s, "/users", option.TagInfo("user", "routes relative to users"))
	fuego.Get(userGroup, "/{publicKey}", GetUser(chatService))
}
