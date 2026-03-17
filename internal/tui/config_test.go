package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveLoadConfig_Roundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	cfg := appConfig{
		Identity: &identityConfig{
			PrivateKey: id.PrivKeyHex,
			PublicKey:  id.PubKeyHex,
		},
		Servers: []serverConfig{
			{URL: "http://server1.example", Quickname: "S1", Description: "First"},
			{URL: "http://server2.example", Quickname: "S2", Description: "Second"},
		},
		LastServer: "http://server1.example",
	}

	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	if got.Identity == nil {
		t.Fatal("Identity should not be nil after roundtrip")
	}
	if got.Identity.PrivateKey != cfg.Identity.PrivateKey {
		t.Errorf("Identity.PrivateKey = %q, want %q", got.Identity.PrivateKey, cfg.Identity.PrivateKey)
	}
	if got.Identity.PublicKey != cfg.Identity.PublicKey {
		t.Errorf("Identity.PublicKey = %q, want %q", got.Identity.PublicKey, cfg.Identity.PublicKey)
	}
	if got.LastServer != cfg.LastServer {
		t.Errorf("LastServer = %q, want %q", got.LastServer, cfg.LastServer)
	}
	if len(got.Servers) != len(cfg.Servers) {
		t.Fatalf("Servers count = %d, want %d", len(got.Servers), len(cfg.Servers))
	}
	for i, s := range got.Servers {
		want := cfg.Servers[i]
		if s.URL != want.URL {
			t.Errorf("Servers[%d].URL = %q, want %q", i, s.URL, want.URL)
		}
		if s.Quickname != want.Quickname {
			t.Errorf("Servers[%d].Quickname = %q, want %q", i, s.Quickname, want.Quickname)
		}
		if s.Description != want.Description {
			t.Errorf("Servers[%d].Description = %q, want %q", i, s.Description, want.Description)
		}
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig on missing file should not error, got: %v", err)
	}
	if cfg.Identity != nil {
		t.Error("Identity should be nil for missing config")
	}
	if len(cfg.Servers) != 0 {
		t.Errorf("Servers should be empty, got %v", cfg.Servers)
	}
}

func TestSaveConfig_CreatesParentDirs(t *testing.T) {
	// t.Setenv("HOME", t.TempDir()) ensures $HOME has no .config/microchat dir yet
	t.Setenv("HOME", t.TempDir())

	cfg := appConfig{LastServer: "http://test.example"}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig should create parent directories: %v", err)
	}

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.LastServer != cfg.LastServer {
		t.Errorf("LastServer = %q, want %q", got.LastServer, cfg.LastServer)
	}
}

func TestCheckConfigPermissions_MissingFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := checkConfigPermissions(); err != nil {
		t.Errorf("expected nil for missing file, got: %v", err)
	}
}

func TestCheckConfigPermissions_CorrectPerms(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := saveConfig(appConfig{}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	if err := checkConfigPermissions(); err != nil {
		t.Errorf("expected nil for correct perms, got: %v", err)
	}
}

func TestCheckConfigPermissions_WrongFilePerms(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := saveConfig(appConfig{}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	if err := os.Chmod(configPath(), 0644); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	err := checkConfigPermissions()
	if err == nil {
		t.Error("expected error for wrong file permissions")
	}
	if err != nil && !strings.Contains(err.Error(), "0644") {
		t.Errorf("error should mention wrong permissions, got: %v", err)
	}
}

func TestCheckConfigPermissions_WrongDirPerms(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := saveConfig(appConfig{}); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	dirPath := filepath.Dir(configPath())
	if err := os.Chmod(dirPath, 0755); err != nil {
		t.Fatalf("chmod dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dirPath, 0700) })
	err := checkConfigPermissions()
	if err == nil {
		t.Error("expected error for wrong dir permissions")
	}
	if err != nil && !strings.Contains(err.Error(), "0755") {
		t.Errorf("error should mention wrong permissions, got: %v", err)
	}
}

func TestSaveLoadConfig_WithUsers(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := appConfig{
		Users: []userEntry{
			{PubKey: "abc123", DisplayName: "Alice"},
			{PubKey: "def456", DisplayName: "Bob"},
		},
	}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(got.Users) != 2 {
		t.Fatalf("Users count = %d, want 2", len(got.Users))
	}
	if got.Users[0].PubKey != "abc123" {
		t.Errorf("Users[0].PubKey = %q, want abc123", got.Users[0].PubKey)
	}
	if got.Users[0].DisplayName != "Alice" {
		t.Errorf("Users[0].DisplayName = %q, want Alice", got.Users[0].DisplayName)
	}
	if got.Users[1].PubKey != "def456" {
		t.Errorf("Users[1].PubKey = %q, want def456", got.Users[1].PubKey)
	}
}

func TestSaveLoadConfig_EmptyConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg := appConfig{}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.Identity != nil {
		t.Error("Identity should be nil")
	}
	if len(got.Servers) != 0 {
		t.Error("Servers should be empty")
	}
}
