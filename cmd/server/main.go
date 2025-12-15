package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/EwenQuim/microchat/internal/handlers"
	"github.com/EwenQuim/microchat/internal/repository/memory"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

func main() {
	// Initialize repository
	repo := memory.NewStore()

	// Initialize services
	chatService := services.NewChatService(repo)

	// Create Fuego server with port
	s := fuego.NewServer(
		fuego.WithAddr("0.0.0.0:9997"),
	)

	// API routes
	apiGroup := fuego.Group(s, "/api")
	handlers.RegisterChatRoutes(apiGroup, chatService)

	// Serve static files with SPA fallback
	spaHandler := createSPAHandler("./static")
	fuego.GetStd(s, "/", spaHandler)
	fuego.GetStd(s, "/*", spaHandler)

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}

// createSPAHandler creates a handler that serves static files and falls back to index.html for SPA routes
func createSPAHandler(staticDir string) http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(staticDir))

	assetExtensions := []string{".js", ".css", ".json", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}

	return func(w http.ResponseWriter, r *http.Request) {
		// Build the full file path
		path := filepath.Join(staticDir, r.URL.Path)

		// Check if the file exists
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			// Check if this is a request for an asset file (should return 404)
			ext := filepath.Ext(r.URL.Path)
			if slices.Contains(assetExtensions, ext) {
				http.NotFound(w, r)
				return
			}

			// Not an asset file, serve index.html for SPA routing
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}

		// File exists, serve it
		fileServer.ServeHTTP(w, r)
	}
}
