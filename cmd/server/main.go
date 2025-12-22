package main

import (
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/EwenQuim/microchat/internal/handlers"
	"github.com/EwenQuim/microchat/internal/repository"
	"github.com/EwenQuim/microchat/internal/services"

	"github.com/go-fuego/fuego"
)

//go:embed all:static
var staticFiles embed.FS

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize repository
	repo, err := repository.NewRepository()
	if err != nil {
		slog.Error("Failed to initialize repository", "error", err)
		os.Exit(1)
	}

	// Initialize services
	chatService := services.NewChatService(repo)

	// Create Fuego server with port
	s := fuego.NewServer(
		fuego.WithAddr("0.0.0.0:9997"),
		fuego.WithEngineOptions(
			fuego.WithOpenAPIConfig(fuego.OpenAPIConfig{
				PrettyFormatJSON: true,
				Disabled:         os.Getenv("ENV") != "dev",
			}),
		),
	)

	// API routes
	apiGroup := fuego.Group(s, "/api")
	handlers.RegisterChatRoutes(apiGroup, chatService, cfg)

	// Serve static files with SPA fallback
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		slog.Error("Failed to create sub filesystem", "error", err)
		os.Exit(1)
	}
	spaHandler := createSPAHandler(staticFS)
	fuego.GetStd(s, "/", spaHandler)
	fuego.GetStd(s, "/*", spaHandler)

	if err := s.Run(); err != nil {
		slog.Error("Server failed to run", "error", err)
		os.Exit(1)
	}
}

// createSPAHandler creates a handler that serves static files and falls back to index.html for SPA routes
func createSPAHandler(staticFS fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFS))

	assetExtensions := []string{".js", ".css", ".json", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot"}

	return func(w http.ResponseWriter, r *http.Request) {
		// Clean the path to prevent directory traversal
		path := filepath.Clean(r.URL.Path)
		if path == "/" {
			path = "index.html"
		} else {
			// Remove leading slash for fs.FS
			path = path[1:]
		}

		// Check if the file exists in the embedded FS
		_, err := fs.Stat(staticFS, path)
		if err != nil {
			// Check if this is a request for an asset file (should return 404)
			ext := filepath.Ext(r.URL.Path)
			if slices.Contains(assetExtensions, ext) {
				http.NotFound(w, r)
				return
			}

			// Not an asset file, serve index.html for SPA routing
			indexData, err := fs.ReadFile(staticFS, "index.html")
			if err != nil {
				http.Error(w, "index.html not found", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, err = w.Write(indexData)
			if err != nil {
				slog.Error("Cannot write static response", "err", err)
			}
			return
		}

		// File exists, serve it
		fileServer.ServeHTTP(w, r)
	}
}
