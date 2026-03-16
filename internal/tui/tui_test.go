package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestInitialModel_NoIdentity_StartsOnIdentityScreen(t *testing.T) {
	cfg := appConfig{}
	m := initialModel(cfg)

	if m.screen != screenIdentity {
		t.Errorf("screen = %v, want screenIdentity", m.screen)
	}
	if m.id != nil {
		t.Error("id should be nil when no identity in config")
	}
}

func TestInitialModel_WithIdentity_StartsOnServersScreen(t *testing.T) {
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
	}
	m := initialModel(cfg)

	if m.screen != screenServers {
		t.Errorf("screen = %v, want screenServers", m.screen)
	}
	if m.id == nil {
		t.Error("id should be set when identity is in config")
	}
	if m.id.PubKeyHex != id.PubKeyHex {
		t.Errorf("id.PubKeyHex = %q, want %q", m.id.PubKeyHex, id.PubKeyHex)
	}
}

func TestInitialModel_InvalidPrivKey_StartsOnServersScreen(t *testing.T) {
	cfg := appConfig{
		Identity: &identityConfig{
			PrivateKey: "not-valid-hex!!",
			PublicKey:  "",
		},
	}
	m := initialModel(cfg)

	// Should still go to servers screen (identity was configured)
	if m.screen != screenServers {
		t.Errorf("screen = %v, want screenServers", m.screen)
	}
	// But id should be nil due to parse error
	if m.id != nil {
		t.Error("id should be nil when private key is invalid")
	}
}

func TestModel_WindowSizeMsg(t *testing.T) {
	m := initialModel(appConfig{})
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := newModel.(model)

	if updated.width != 120 {
		t.Errorf("width = %d, want 120", updated.width)
	}
	if updated.height != 40 {
		t.Errorf("height = %d, want 40", updated.height)
	}
}

func TestModel_NavigateMsg_ToIdentity(t *testing.T) {
	m := initialModel(appConfig{})
	newModel, _ := m.Update(navigateMsg{to: screenIdentity})
	updated := newModel.(model)

	if updated.screen != screenIdentity {
		t.Errorf("screen = %v, want screenIdentity", updated.screen)
	}
}

func TestModel_NavigateMsg_ToServers_SavesIdentity(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	m := initialModel(appConfig{})
	// Simulate identity screen completing — set ident.result
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	m.screen = screenIdentity
	m.ident.result = id

	newModel, _ := m.Update(navigateMsg{to: screenServers})
	updated := newModel.(model)

	if updated.screen != screenServers {
		t.Errorf("screen = %v, want screenServers", updated.screen)
	}
	if updated.id == nil {
		t.Error("id should be saved after navigating from identity to servers")
	}
	if updated.id.PubKeyHex != id.PubKeyHex {
		t.Errorf("id.PubKeyHex = %q, want %q", updated.id.PubKeyHex, id.PubKeyHex)
	}
	if updated.cfg.Identity == nil {
		t.Error("cfg.Identity should be set after navigating from identity")
	}
}

func TestModel_View_DelegatesToIdentityScreen(t *testing.T) {
	m := initialModel(appConfig{})
	m.screen = screenIdentity
	m.width = 80
	m.height = 24

	v := m.View()
	if !strings.Contains(v.Content, "[g]") {
		t.Errorf("view should contain identity screen content, got:\n%s", v.Content)
	}
}

func TestModel_View_DelegatesToServersScreen(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	cfg := appConfig{
		Identity: &identityConfig{
			PrivateKey: id.PrivKeyHex,
			PublicKey:  id.PubKeyHex,
		},
	}
	m := initialModel(cfg)
	m.width = 80
	m.height = 24

	v := m.View()
	if !strings.Contains(v.Content, "Servers") {
		t.Errorf("view should contain servers screen content, got:\n%s", v.Content)
	}
}
