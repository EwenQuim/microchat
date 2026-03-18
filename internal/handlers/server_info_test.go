package handlers

import (
	"testing"

	"github.com/EwenQuim/microchat/internal/config"
	"github.com/go-fuego/fuego"
)

func TestGetServerInfo_NoSuggestedServers(t *testing.T) {
	cfg := &config.Config{
		QuickName:   "testserver",
		Description: "A test server",
	}
	handler := GetServerInfo(cfg)
	resp, err := handler(fuego.NewMockContextNoBody())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SuggestedServers != nil {
		t.Errorf("expected nil SuggestedServers, got %v", resp.SuggestedServers)
	}
}

func TestGetServerInfo_WithSuggestedServers(t *testing.T) {
	cfg := &config.Config{
		QuickName:   "testserver",
		Description: "A test server",
		SuggestedServerList: []string{
			"https://backup.example.com",
			"https://other.example.com",
		},
	}
	handler := GetServerInfo(cfg)
	resp, err := handler(fuego.NewMockContextNoBody())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.SuggestedServers) != 2 {
		t.Fatalf("expected 2 suggested servers, got %d: %v", len(resp.SuggestedServers), resp.SuggestedServers)
	}
	if resp.SuggestedServers[0] != "https://backup.example.com" {
		t.Errorf("SuggestedServers[0] = %q, want %q", resp.SuggestedServers[0], "https://backup.example.com")
	}
	if resp.SuggestedServers[1] != "https://other.example.com" {
		t.Errorf("SuggestedServers[1] = %q, want %q", resp.SuggestedServers[1], "https://other.example.com")
	}
}
