package tui

import (
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
