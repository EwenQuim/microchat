package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func makeIdentitiesModel(active int, entries ...identityEntry) identitiesModel {
	cfg := appConfig{Identities: entries, ActiveIndex: active}
	return newIdentitiesModel(cfg)
}

func TestIdentitiesModel_NewFromConfig(t *testing.T) {
	id, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{Name: "Main", PrivateKey: id.PrivKeyHex, PublicKey: id.PubKeyHex})
	if len(m.entries) != 1 {
		t.Errorf("entries count = %d, want 1", len(m.entries))
	}
	if m.activeIndex != 0 {
		t.Errorf("activeIndex = %d, want 0", m.activeIndex)
	}
}

func TestIdentitiesModel_CursorDown(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0,
		identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m2.cursor)
	}
}

func TestIdentitiesModel_CursorUp(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0,
		identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	m.cursor = 1
	m2, _ := m.update(pressKey(tea.KeyUp))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m2.cursor)
	}
}

func TestIdentitiesModel_CursorDown_j(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0,
		identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	m2, _ := m.update(pressChar("j"))
	if m2.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m2.cursor)
	}
}

func TestIdentitiesModel_CursorUp_k(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0,
		identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	m.cursor = 1
	m2, _ := m.update(pressChar("k"))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m2.cursor)
	}
}

func TestIdentitiesModel_Enter_SetsActive(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0,
		identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	m.cursor = 1
	m2, _ := m.update(pressKey(tea.KeyEnter))
	if m2.activeIndex != 1 {
		t.Errorf("activeIndex = %d, want 1", m2.activeIndex)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true after switching active")
	}
}

func TestIdentitiesModel_Delete_NonLast(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0,
		identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	m.cursor = 0
	m2, _ := m.update(pressChar("d"))
	if len(m2.entries) != 1 {
		t.Errorf("entries count = %d, want 1", len(m2.entries))
	}
	if !m2.configChanged {
		t.Error("configChanged should be true after delete")
	}
}

func TestIdentitiesModel_Delete_Last_NoOp(t *testing.T) {
	id1, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	m2, _ := m.update(pressChar("d"))
	if len(m2.entries) != 1 {
		t.Errorf("entries count = %d, want 1 (cannot delete last)", len(m2.entries))
	}
}

func TestIdentitiesModel_Delete_AdjustsActiveIndex(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(1,
		identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	m.cursor = 0
	m2, _ := m.update(pressChar("d"))
	if m2.activeIndex != 0 {
		t.Errorf("activeIndex = %d, want 0 after deleting entry before active", m2.activeIndex)
	}
}

func TestIdentitiesModel_Esc_NavigatesToRooms(t *testing.T) {
	id1, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	_, cmd := m.update(pressKey(tea.KeyEscape))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenRooms {
		t.Errorf("expected screenRooms, got %v", nav.to)
	}
}

func TestIdentitiesModel_Tab_NavigatesToContacts(t *testing.T) {
	id1, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	_, cmd := m.update(pressKey(tea.KeyTab))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenContacts {
		t.Errorf("expected screenContacts, got %v", nav.to)
	}
}

func TestIdentitiesModel_PressA_EntersAddKeyState(t *testing.T) {
	id1, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	m2, _ := m.update(pressChar("a"))
	if m2.state != identitiesStateAddKey {
		t.Errorf("state = %v, want identitiesStateAddKey", m2.state)
	}
}

func TestIdentitiesModel_AddKey_EscAtMenuCancels(t *testing.T) {
	id1, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	m.state = identitiesStateAddKey
	m.sub = newIdentityModelAdd()
	// sub is in idStateMenu, pressing esc should cancel
	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != identitiesStateList {
		t.Errorf("state = %v, want identitiesStateList", m2.state)
	}
}

func TestIdentitiesModel_AddKey_IdentityCreated_TransitionsToAddName(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	m.state = identitiesStateAddKey
	m.sub = newIdentityModelAdd()
	m2, _ := m.update(identityCreatedMsg{id: id2})
	if m2.state != identitiesStateAddName {
		t.Errorf("state = %v, want identitiesStateAddName", m2.state)
	}
	if m2.pendingKey.PubKeyHex != id2.PubKeyHex {
		t.Errorf("pendingKey not set correctly")
	}
}

func TestIdentitiesModel_AddName_EnterAddsEntry(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	m.state = identitiesStateAddName
	m.pendingKey = id2
	m.inputName = "My Alt Key"
	m2, _ := m.update(pressKey(tea.KeyEnter))
	if len(m2.entries) != 2 {
		t.Errorf("entries count = %d, want 2", len(m2.entries))
	}
	if m2.entries[1].Name != "My Alt Key" {
		t.Errorf("entries[1].Name = %q, want My Alt Key", m2.entries[1].Name)
	}
	if m2.entries[1].PrivateKey != id2.PrivKeyHex {
		t.Errorf("entries[1].PrivateKey not set correctly")
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
	if m2.state != identitiesStateList {
		t.Errorf("state = %v, want identitiesStateList", m2.state)
	}
}

func TestIdentitiesModel_AddName_EscCancels(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	m.state = identitiesStateAddName
	m.pendingKey = id2
	m.inputName = "Typed"
	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != identitiesStateList {
		t.Errorf("state = %v, want identitiesStateList", m2.state)
	}
	if len(m2.entries) != 1 {
		t.Errorf("entries count = %d, want 1 (no entry added on esc)", len(m2.entries))
	}
}

func TestIdentitiesModel_View_ShowsActiveMarker(t *testing.T) {
	id1, _ := generateIdentity()
	id2, _ := generateIdentity()
	m := makeIdentitiesModel(0,
		identityEntry{Name: "Main", PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex},
		identityEntry{Name: "Alt", PrivateKey: id2.PrivKeyHex, PublicKey: id2.PubKeyHex},
	)
	v := m.view(80, 24)
	if !strings.Contains(v, "*") {
		t.Error("view should show * for active identity")
	}
}

func TestIdentitiesModel_View_ShowsCursor(t *testing.T) {
	id1, _ := generateIdentity()
	m := makeIdentitiesModel(0, identityEntry{Name: "Main", PrivateKey: id1.PrivKeyHex, PublicKey: id1.PubKeyHex})
	v := m.view(80, 24)
	if !strings.Contains(v, ">") {
		t.Error("view should show cursor indicator for selected row")
	}
}
