package config

import (
	"testing"
)

func TestLoad_SuggestedServerList_Empty(t *testing.T) {
	t.Setenv("SUGGESTED_SERVER_LIST", "")
	cfg := Load()
	if len(cfg.SuggestedServerList) != 0 {
		t.Errorf("expected empty slice, got %v", cfg.SuggestedServerList)
	}
}

func TestLoad_SuggestedServerList_SingleURL(t *testing.T) {
	t.Setenv("SUGGESTED_SERVER_LIST", "https://backup.example.com")
	cfg := Load()
	if len(cfg.SuggestedServerList) != 1 {
		t.Fatalf("expected 1 entry, got %d: %v", len(cfg.SuggestedServerList), cfg.SuggestedServerList)
	}
	if cfg.SuggestedServerList[0] != "https://backup.example.com" {
		t.Errorf("entry = %q, want %q", cfg.SuggestedServerList[0], "https://backup.example.com")
	}
}

func TestLoad_SuggestedServerList_MultipleURLs(t *testing.T) {
	t.Setenv("SUGGESTED_SERVER_LIST", "https://backup.example.com , https://other.example.com")
	cfg := Load()
	if len(cfg.SuggestedServerList) != 2 {
		t.Fatalf("expected 2 entries, got %d: %v", len(cfg.SuggestedServerList), cfg.SuggestedServerList)
	}
	if cfg.SuggestedServerList[0] != "https://backup.example.com" {
		t.Errorf("entry[0] = %q, want %q", cfg.SuggestedServerList[0], "https://backup.example.com")
	}
	if cfg.SuggestedServerList[1] != "https://other.example.com" {
		t.Errorf("entry[1] = %q, want %q", cfg.SuggestedServerList[1], "https://other.example.com")
	}
}
