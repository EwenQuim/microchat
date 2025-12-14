package main

import (
	"log"
	"net/http"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/handlers"
	"github.com/EwenQuim/microchat/internal/repository/memory"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func main() {
	cfg := config.Load()

	// Initialize repository
	repo := memory.NewStore()

	// Initialize services
	chatService := services.NewChatService(repo)

	// Create Fuego server with port
	s := fuego.NewServer(
		fuego.WithAddr("0.0.0.0:9999"),
	)

	// API routes
	apiGroup := fuego.Group(s, "/api")
	handlers.RegisterChatRoutes(apiGroup, chatService)

	// Serve static files (frontend)
	fs := http.FileServer(http.Dir("./static"))
	fuego.GetStd(s, "/", fs.ServeHTTP)
	fuego.GetStd(s, "/*", fs.ServeHTTP)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
