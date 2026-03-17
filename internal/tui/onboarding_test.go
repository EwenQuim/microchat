package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// TestNewUser_NoConfig_SeesIdentityScreen verifies that a fresh install with no
// config file starts on the identity setup screen with no identity loaded.
func TestNewUser_NoConfig_SeesIdentityScreen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}

	m := initialModel(cfg)

	if m.screen != screenIdentity {
		t.Errorf("screen = %v, want screenIdentity", m.screen)
	}
	if m.id != nil {
		t.Error("id should be nil for a new user with no config")
	}
}

// TestNewUser_FreshStart_IdentityScreenShowsOptions verifies that the identity
// screen shown on first launch presents all expected options and no errors.
func TestNewUser_FreshStart_IdentityScreenShowsOptions(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, _ := loadConfig()
	m := initialModel(cfg)
	m.width = 80
	m.height = 24

	v := m.View()
	for _, want := range []string{"g ", "p ", "q "} {
		if !strings.Contains(v.Content, want) {
			t.Errorf("identity screen view missing %q option; got:\n%s", want, v.Content)
		}
	}
	if strings.Contains(v.Content, "Error") {
		t.Errorf("fresh identity screen should show no errors; got:\n%s", v.Content)
	}
}

// TestNewUser_GenerateKey_NavigationSavesConfig verifies the full generate-key
// onboarding path: press g → navigate cmd → dispatch → config file created.
func TestNewUser_GenerateKey_NavigationSavesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, _ := loadConfig()
	m := initialModel(cfg)

	// Press g on the identity screen
	newModel, cmd := m.Update(pressChar("g"))
	m = newModel.(model)

	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg after pressing g, got %T", msg)
	}
	if nav.to != screenServers {
		t.Errorf("nav.to = %v, want screenServers", nav.to)
	}

	// Dispatch the navigate message (this triggers saveConfig)
	newModel, _ = m.Update(nav)
	m = newModel.(model)

	if m.screen != screenServers {
		t.Errorf("screen = %v, want screenServers after navigate", m.screen)
	}

	saved, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig after onboarding: %v", err)
	}
	if len(saved.Identities) == 0 {
		t.Error("config should contain at least one identity after onboarding")
	}
}

// TestNewUser_AfterOnboarding_RestartSeesServersScreen simulates loading the
// saved config on a second startup and verifies the user lands on the servers screen.
func TestNewUser_AfterOnboarding_RestartSeesServersScreen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// First startup: generate key and navigate
	cfg, _ := loadConfig()
	m := initialModel(cfg)
	newModel, cmd := m.Update(pressChar("g"))
	m = newModel.(model)
	msg := runCmd(cmd)
	newModel, _ = m.Update(msg)
	m = newModel.(model)

	if m.screen != screenServers {
		t.Fatalf("expected screenServers after first onboarding, got %v", m.screen)
	}

	// Simulate restart: load config and build a fresh model
	cfg2, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig on restart: %v", err)
	}
	m2 := initialModel(cfg2)

	if m2.screen != screenServers {
		t.Errorf("after restart, screen = %v, want screenServers", m2.screen)
	}
	if m2.id == nil {
		t.Error("after restart, id should be set from saved config")
	}
}

// TestNewUser_PasteKey_NavigationSavesConfig verifies the paste-key onboarding
// path: p → type key → Enter → navigate → config persisted with correct key.
func TestNewUser_PasteKey_NavigationSavesConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	cfg, _ := loadConfig()
	m := initialModel(cfg)

	// Enter paste mode
	newModel, _ := m.Update(pressChar("p"))
	m = newModel.(model)

	// Inject the private key into the input buffer directly (simulates typing)
	m.ident.inputText = id.PrivKeyHex

	// Confirm with Enter
	newModel, cmd := m.Update(pressKey(tea.KeyEnter))
	m = newModel.(model)

	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg after entering valid key, got %T", msg)
	}

	// Dispatch navigate
	newModel, _ = m.Update(nav)
	m = newModel.(model)

	saved, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(saved.Identities) == 0 {
		t.Fatal("config should have at least one identity after paste onboarding")
	}
	if saved.Identities[0].PublicKey != id.PubKeyHex {
		t.Errorf("saved PublicKey = %q, want %q", saved.Identities[0].PublicKey, id.PubKeyHex)
	}
}

// TestNewUser_ConfigFormat_AfterOnboarding verifies that after generating a key
// the config uses Identities (not legacy Identity), ActiveIndex is 0, and
// Identity is nil.
func TestNewUser_ConfigFormat_AfterOnboarding(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, _ := loadConfig()
	m := initialModel(cfg)
	newModel, cmd := m.Update(pressChar("g"))
	m = newModel.(model)
	nav := runCmd(cmd).(navigateMsg)
	m.Update(nav) //nolint:errcheck

	saved, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(saved.Identities) != 1 {
		t.Errorf("Identities count = %d, want 1", len(saved.Identities))
	}
	if saved.ActiveIndex != 0 {
		t.Errorf("ActiveIndex = %d, want 0", saved.ActiveIndex)
	}
	if saved.Identity != nil {
		t.Error("legacy Identity field should be nil after onboarding via Identities slice")
	}
}

// TestNewUser_ConfigPermissions_AreSecure verifies that the config file and
// directory created during onboarding have secure permissions (0600 / 0700).
func TestNewUser_ConfigPermissions_AreSecure(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, _ := loadConfig()
	m := initialModel(cfg)
	newModel, cmd := m.Update(pressChar("g"))
	m = newModel.(model)
	nav := runCmd(cmd).(navigateMsg)
	m.Update(nav) //nolint:errcheck

	path := configPath()

	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("config file not found after onboarding: %v", err)
	}
	if got := fileInfo.Mode().Perm(); got != 0600 {
		t.Errorf("config file permissions = %04o, want 0600", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("config dir not found: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0700 {
		t.Errorf("config dir permissions = %04o, want 0700", got)
	}
}

// TestNewUser_MultipleStartups_NoIdentityDuplication generates two distinct
// keys across two onboarding cycles and verifies the config accumulates exactly
// 2 entries without duplicating existing ones.
func TestNewUser_MultipleStartups_NoIdentityDuplication(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	// First cycle: generate key #1 and navigate to servers
	cfg, _ := loadConfig()
	m := initialModel(cfg)
	newModel, cmd := m.Update(pressChar("g"))
	m = newModel.(model)
	msg := runCmd(cmd)
	newModel, _ = m.Update(msg)
	m = newModel.(model)

	if m.screen != screenServers {
		t.Fatalf("expected screenServers after first cycle, got %v", m.screen)
	}

	// Navigate back to identity screen
	newModel, _ = m.Update(navigateMsg{to: screenIdentity})
	m = newModel.(model)

	// Second cycle: generate key #2 and navigate to servers
	newModel, cmd = m.Update(pressChar("g"))
	m = newModel.(model)
	msg = runCmd(cmd)
	newModel, _ = m.Update(msg)
	m = newModel.(model)

	saved, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(saved.Identities) != 2 {
		t.Errorf("expected exactly 2 identities after two generate cycles, got %d", len(saved.Identities))
	}
}
