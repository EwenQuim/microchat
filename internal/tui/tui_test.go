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

func TestInitialModel_WithIdentity_StartsOnRoomsScreen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	// Uses legacy Identity field - should still work via fallback
	cfg := appConfig{
		Identity: &identityConfig{
			PrivateKey: id.PrivKeyHex,
			PublicKey:  id.PubKeyHex,
		},
	}
	m := initialModel(cfg)

	if m.screen != screenRooms {
		t.Errorf("screen = %v, want screenRooms", m.screen)
	}
	if m.id == nil {
		t.Error("id should be set when identity is in config")
	}
	if m.id.PubKeyHex != id.PubKeyHex {
		t.Errorf("id.PubKeyHex = %q, want %q", m.id.PubKeyHex, id.PubKeyHex)
	}
}

func TestInitialModel_InvalidPrivKey_StartsOnRoomsScreen(t *testing.T) {
	cfg := appConfig{
		Identity: &identityConfig{
			PrivateKey: "not-valid-hex!!",
			PublicKey:  "",
		},
	}
	m := initialModel(cfg)

	// Should still go to rooms screen (identity was configured)
	if m.screen != screenRooms {
		t.Errorf("screen = %v, want screenRooms", m.screen)
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
	if len(updated.cfg.Identities) == 0 {
		t.Error("cfg.Identities should be set after navigating from identity")
	}
}

func TestModel_View_DelegatesToIdentityScreen(t *testing.T) {
	m := initialModel(appConfig{})
	m.screen = screenIdentity
	m.width = 80
	m.height = 24

	v := m.View()
	if !strings.Contains(v.Content, "generate new keypair") {
		t.Errorf("view should contain identity screen content, got:\n%s", v.Content)
	}
}

func TestModel_NavigateMsg_ToContacts(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
		},
		Contacts: []contactEntry{
			{PubKey: "abc", DisplayName: "Alice"},
		},
	}
	m := initialModel(cfg)
	newModel, _ := m.Update(navigateMsg{to: screenContacts})
	updated := newModel.(model)
	if updated.screen != screenContacts {
		t.Errorf("screen = %v, want screenContacts", updated.screen)
	}
	if len(updated.contacts.contacts) != 1 {
		t.Errorf("contacts count = %d, want 1", len(updated.contacts.contacts))
	}
}

func TestModel_View_DelegatesToContactsScreen(t *testing.T) {
	m := initialModel(appConfig{})
	m.screen = screenContacts
	m.contacts = newContactsModel(appConfig{})
	m.width = 80
	m.height = 24
	v := m.View()
	if !strings.Contains(v.Content, "Contacts") {
		t.Errorf("view should contain contacts screen content, got:\n%s", v.Content)
	}
}

func TestModel_ContactsScreen_ConfigSavedOnDelete(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := initialModel(appConfig{})
	m.screen = screenContacts
	m.contacts = newContactsModel(appConfig{
		Contacts: []contactEntry{{PubKey: "abc", DisplayName: "Alice"}},
	})
	newModel, _ := m.Update(pressChar("d"))
	updated := newModel.(model)
	if updated.contacts.configChanged {
		t.Error("configChanged should be reset after save")
	}
	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if len(got.Contacts) != 0 {
		t.Errorf("expected 0 contacts in saved config, got %d", len(got.Contacts))
	}
}

func TestModel_View_DelegatesToServersScreen(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
		},
	}
	m := initialModel(cfg)
	// initialModel now starts on screenRooms; manually switch to servers for this view test
	m.screen = screenServers
	m.width = 80
	m.height = 24

	v := m.View()
	if !strings.Contains(v.Content, "Servers") {
		t.Errorf("view should contain servers screen content, got:\n%s", v.Content)
	}
}

func TestModel_NavigateMsg_ToIdentities(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
		},
	}
	m := initialModel(cfg)
	newModel, _ := m.Update(navigateMsg{to: screenIdentities})
	updated := newModel.(model)
	if updated.screen != screenIdentities {
		t.Errorf("screen = %v, want screenIdentities", updated.screen)
	}
}

func TestModel_View_DelegatesToIdentitiesScreen(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	cfg := appConfig{
		Identities: []identityEntry{
			{Name: "Main", PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
		},
	}
	m := initialModel(cfg)
	m.screen = screenIdentities
	m.identities = newIdentitiesModel(m.cfg)
	m.width = 80
	m.height = 24
	v := m.View()
	if !strings.Contains(v.Content, "Identities") {
		t.Errorf("view should contain identities screen content, got:\n%s", v.Content)
	}
}

func TestInitialModel_WithIdentitiesArray_StartsOnRoomsScreen(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
		},
	}
	m := initialModel(cfg)

	if m.screen != screenRooms {
		t.Errorf("screen = %v, want screenRooms", m.screen)
	}
	if m.id == nil {
		t.Fatal("id should be set when Identities array is populated")
	}
	if m.id.PubKeyHex != id.PubKeyHex {
		t.Errorf("id.PubKeyHex = %q, want %q", m.id.PubKeyHex, id.PubKeyHex)
	}
}

func TestModel_AddContactFromChat_SavedToConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
		},
	}
	m := initialModel(cfg)
	m.screen = screenRooms

	newModel, _ := m.Update(addContactFromChatMsg{pubKeyHex: "deadbeef", displayName: "Bob"})
	updated := newModel.(model)

	if len(updated.cfg.Contacts) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(updated.cfg.Contacts))
	}
	if updated.cfg.Contacts[0].PubKey != "deadbeef" {
		t.Errorf("PubKey = %q, want %q", updated.cfg.Contacts[0].PubKey, "deadbeef")
	}
	if updated.cfg.Contacts[0].DisplayName != "Bob" {
		t.Errorf("DisplayName = %q, want %q", updated.cfg.Contacts[0].DisplayName, "Bob")
	}
}

func TestModel_AddContactFromChat_Deduplicates(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex},
		},
	}
	m := initialModel(cfg)
	m.screen = screenRooms

	m2, _ := m.Update(addContactFromChatMsg{pubKeyHex: "deadbeef", displayName: "Bob"})
	m3, _ := m2.(model).Update(addContactFromChatMsg{pubKeyHex: "deadbeef", displayName: "Bobby"})
	updated := m3.(model)

	if len(updated.cfg.Contacts) != 1 {
		t.Fatalf("expected 1 contact after dedup, got %d", len(updated.cfg.Contacts))
	}
	if updated.cfg.Contacts[0].DisplayName != "Bobby" {
		t.Errorf("DisplayName = %q, want %q (should be updated)", updated.cfg.Contacts[0].DisplayName, "Bobby")
	}
}

func TestInitialModel_IdentitiesArray_ActiveIndexOutOfBounds_FallsToFirst(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	cfg := appConfig{
		Identities:  []identityEntry{{PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex}},
		ActiveIndex: 99,
	}
	m := initialModel(cfg)

	if m.id == nil {
		t.Fatal("id should be set (falls back to index 0)")
	}
	if m.id.PubKeyHex != id.PubKeyHex {
		t.Errorf("id.PubKeyHex = %q, want %q (should fall back to first entry)", m.id.PubKeyHex, id.PubKeyHex)
	}
}

func TestInitialModel_BothIdentitiesAndLegacy_IdentitiesWins(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id1, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity id1: %v", err)
	}
	id2, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity id2: %v", err)
	}

	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		},
		Identity: &identityConfig{
			PrivateKey: id2.PrivKeyHex,
			PublicKey:  id2.PubKeyHex,
		},
	}
	m := initialModel(cfg)

	if m.id == nil {
		t.Fatal("id should be set")
	}
	if m.id.PubKeyHex != id1.PubKeyHex {
		t.Errorf("expected Identities array to win over legacy Identity, got pubkey %q, want %q", m.id.PubKeyHex, id1.PubKeyHex)
	}
}

func TestModel_NavigateToServers_SavesIdentityToIdentitiesSlice(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	m := initialModel(appConfig{})
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	m.screen = screenIdentity
	m.ident.result = id

	newModel, _ := m.Update(navigateMsg{to: screenServers})
	updated := newModel.(model)

	if len(updated.cfg.Identities) != 1 {
		t.Errorf("cfg.Identities count = %d, want 1", len(updated.cfg.Identities))
	}
	if updated.cfg.Identity != nil {
		t.Error("legacy cfg.Identity should be nil after saving via Identities slice")
	}
}

func TestModel_Identities_ConfigChanged_PersistsAndUpdatesActiveId(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	id1, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity id1: %v", err)
	}
	id2, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity id2: %v", err)
	}
	cfg := appConfig{
		Identities: []identityEntry{
			{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
			{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
		},
		ActiveIndex: 0,
	}

	m := initialModel(cfg)
	m.screen = screenIdentities
	m.identities = newIdentitiesModel(cfg)
	m.identities.cursor = 1 // point at second entry

	newModel, _ := m.Update(pressKey(tea.KeyEnter))
	updated := newModel.(model)

	if updated.identities.configChanged {
		t.Error("configChanged should be reset after save")
	}
	if updated.cfg.ActiveIndex != 1 {
		t.Errorf("cfg.ActiveIndex = %d, want 1", updated.cfg.ActiveIndex)
	}
	if updated.id == nil {
		t.Fatal("m.id should be updated to the newly activated identity")
	}
	if updated.id.PubKeyHex != id2.PubKeyHex {
		t.Errorf("m.id.PubKeyHex = %q, want %q", updated.id.PubKeyHex, id2.PubKeyHex)
	}

	got, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if got.ActiveIndex != 1 {
		t.Errorf("saved cfg.ActiveIndex = %d, want 1", got.ActiveIndex)
	}
}
