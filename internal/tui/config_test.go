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
		Identities: []identityEntry{
			{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
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

	if len(got.Identities) != 1 {
		t.Fatalf("Identities count = %d, want 1", len(got.Identities))
	}
	if got.Identities[0].PrivateKey != cfg.Identities[0].PrivateKey {
		t.Errorf("Identities[0].PrivateKey = %q, want %q", got.Identities[0].PrivateKey, cfg.Identities[0].PrivateKey)
	}
	if got.Identities[0].PublicKey != cfg.Identities[0].PublicKey {
		t.Errorf("Identities[0].PublicKey = %q, want %q", got.Identities[0].PublicKey, cfg.Identities[0].PublicKey)
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
	if len(cfg.Identities) != 0 {
		t.Error("Identities should be empty for missing config")
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
	if err := os.Chmod(configPath(), 0644); err != nil { //nolint:gosec // intentionally setting insecure perms to test the check
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
	if err := os.Chmod(dirPath, 0755); err != nil { //nolint:gosec // intentionally setting insecure perms to test the check
		t.Fatalf("chmod dir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dirPath, 0700) }) //nolint:gosec // restoring to secure perms in cleanup
	err := checkConfigPermissions()
	if err == nil {
		t.Error("expected error for wrong dir permissions")
	}
	if err != nil && !strings.Contains(err.Error(), "0755") {
		t.Errorf("error should mention wrong permissions, got: %v", err)
	}
}

func TestSaveLoadConfig_WithContacts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	cfg := appConfig{
		Contacts: []contactEntry{
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
	if len(got.Contacts) != 2 {
		t.Fatalf("Contacts count = %d, want 2", len(got.Contacts))
	}
	if got.Contacts[0].PubKey != "abc123" {
		t.Errorf("Contacts[0].PubKey = %q, want abc123", got.Contacts[0].PubKey)
	}
}

func TestLoadConfig_MigratesLegacyIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	// Save old-style config with Identity field
	cfg := appConfig{
		Identity: &identityConfig{
			PrivateKey: id.PrivKeyHex,
			PublicKey:  id.PubKeyHex,
		},
	}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(got.Identities) != 1 {
		t.Fatalf("Identities count = %d, want 1 after migration", len(got.Identities))
	}
	if got.Identities[0].PrivateKey != id.PrivKeyHex {
		t.Errorf("Identities[0].PrivateKey = %q, want %q", got.Identities[0].PrivateKey, id.PrivKeyHex)
	}
	if got.Identity != nil {
		t.Error("Identity should be nil after migration")
	}
}

func TestSaveLoadConfig_WithIdentities(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	cfg := appConfig{
		Identities: []identityEntry{
			{Name: "Main", PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
			{Name: "Alt", PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
		},
		ActiveIndex: 1,
	}
	if err := saveConfig(cfg); err != nil {
		t.Fatalf("saveConfig: %v", err)
	}
	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(got.Identities) != 2 {
		t.Fatalf("Identities count = %d, want 2", len(got.Identities))
	}
	if got.ActiveIndex != 1 {
		t.Errorf("ActiveIndex = %d, want 1", got.ActiveIndex)
	}
	if got.Identities[0].Name != "Main" {
		t.Errorf("Identities[0].Name = %q, want Main", got.Identities[0].Name)
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
	if len(got.Identities) != 0 {
		t.Error("Identities should be empty")
	}
	if len(got.Servers) != 0 {
		t.Error("Servers should be empty")
	}
}
