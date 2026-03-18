package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func makeContactsModel(contacts ...contactEntry) contactsModel {
	return newContactsModel(appConfig{Contacts: contacts})
}

func TestContactsModel_NewFromConfig(t *testing.T) {
	m := makeContactsModel(contactEntry{PubKey: "abc123", DisplayName: "Alice"})
	if len(m.contacts) != 1 {
		t.Errorf("contacts count = %d, want 1", len(m.contacts))
	}
	if m.contacts[0].PubKey != "abc123" {
		t.Errorf("PubKey = %q, want abc123", m.contacts[0].PubKey)
	}
}

func TestContactsModel_CursorUp(t *testing.T) {
	m := makeContactsModel(
		contactEntry{PubKey: "aaa", DisplayName: "Alice"},
		contactEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m.cursor = 1
	m2, _ := m.update(pressKey(tea.KeyUp))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m2.cursor)
	}
}

func TestContactsModel_CursorUp_k(t *testing.T) {
	m := makeContactsModel(
		contactEntry{PubKey: "aaa", DisplayName: "Alice"},
		contactEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m.cursor = 1
	m2, _ := m.update(pressChar("k"))
	if m2.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", m2.cursor)
	}
}

func TestContactsModel_CursorUp_Bounded(t *testing.T) {
	m := makeContactsModel(contactEntry{PubKey: "aaa"})
	m2, _ := m.update(pressKey(tea.KeyUp))
	if m2.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", m2.cursor)
	}
}

func TestContactsModel_CursorDown(t *testing.T) {
	m := makeContactsModel(
		contactEntry{PubKey: "aaa", DisplayName: "Alice"},
		contactEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m2.cursor)
	}
}

func TestContactsModel_CursorDown_j(t *testing.T) {
	m := makeContactsModel(
		contactEntry{PubKey: "aaa", DisplayName: "Alice"},
		contactEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m2, _ := m.update(pressChar("j"))
	if m2.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m2.cursor)
	}
}

func TestContactsModel_CursorDown_Bounded(t *testing.T) {
	m := makeContactsModel(contactEntry{PubKey: "aaa"})
	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 0 {
		t.Errorf("cursor should not exceed len-1, got %d", m2.cursor)
	}
}

func TestContactsModel_Delete(t *testing.T) {
	m := makeContactsModel(
		contactEntry{PubKey: "aaa", DisplayName: "Alice"},
		contactEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m.cursor = 0
	m2, _ := m.update(pressChar("d"))
	if len(m2.contacts) != 1 {
		t.Errorf("contacts count = %d, want 1", len(m2.contacts))
	}
	if m2.contacts[0].PubKey != "bbb" {
		t.Errorf("remaining contact = %q, want bbb", m2.contacts[0].PubKey)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
}

func TestContactsModel_Delete_AdjustsCursor(t *testing.T) {
	m := makeContactsModel(
		contactEntry{PubKey: "aaa"},
		contactEntry{PubKey: "bbb"},
	)
	m.cursor = 1
	m2, _ := m.update(pressChar("d"))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after deleting last item", m2.cursor)
	}
}

func TestContactsModel_Delete_EmptyList(t *testing.T) {
	m := makeContactsModel()
	m2, _ := m.update(pressChar("d"))
	if len(m2.contacts) != 0 {
		t.Errorf("contacts count = %d, want 0", len(m2.contacts))
	}
}

func TestContactsModel_Esc_NavigatesToRooms(t *testing.T) {
	m := makeContactsModel()
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

func TestContactsModel_Tab_NavigatesToRooms(t *testing.T) {
	m := makeContactsModel()
	_, cmd := m.update(pressKey(tea.KeyTab))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenRooms {
		t.Errorf("expected screenRooms, got %v", nav.to)
	}
}

func TestContactsModel_PressA_EntersAddNpubState(t *testing.T) {
	m := makeContactsModel()
	m2, _ := m.update(pressChar("a"))
	if m2.state != contactsStateAddNpub {
		t.Errorf("expected contactsStateAddNpub, got %v", m2.state)
	}
}

func TestContactsModel_AddNpub_TypingBuildsInput(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	m2, _ := m.update(pressChar("n"))
	if m2.inputNpub != "n" {
		t.Errorf("inputNpub = %q, want \"n\"", m2.inputNpub)
	}
}

func TestContactsModel_AddNpub_Backspace(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	m.inputNpub = "ab"
	m2, _ := m.update(pressKey(tea.KeyBackspace))
	if m2.inputNpub != "a" {
		t.Errorf("inputNpub = %q, want \"a\"", m2.inputNpub)
	}
}

func TestContactsModel_AddNpub_EscCancels(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	m.inputNpub = "abc"
	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != contactsStateList {
		t.Errorf("expected contactsStateList, got %v", m2.state)
	}
	if m2.inputNpub != "" {
		t.Errorf("inputNpub should be cleared, got %q", m2.inputNpub)
	}
}

func TestContactsModel_AddNpub_EnterEmpty_Error(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	m2, cmd := m.update(pressKey(tea.KeyEnter))
	if m2.err == "" {
		t.Error("expected error for empty npub")
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty npub")
	}
}

func TestContactsModel_AddNpub_EnterTransitionsToAddName(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	m.inputNpub = "npub1somekey"
	m2, _ := m.update(pressKey(tea.KeyEnter))
	if m2.state != contactsStateAddName {
		t.Errorf("expected contactsStateAddName, got %v", m2.state)
	}
}

func TestContactsModel_AddName_TypingBuildsInput(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddName
	m.inputNpub = "npub1key"
	m2, _ := m.update(pressChar("A"))
	if m2.inputName != "A" {
		t.Errorf("inputName = %q, want \"A\"", m2.inputName)
	}
}

func TestContactsModel_AddName_Backspace(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddName
	m.inputName = "Al"
	m2, _ := m.update(pressKey(tea.KeyBackspace))
	if m2.inputName != "A" {
		t.Errorf("inputName = %q, want \"A\"", m2.inputName)
	}
}

func TestContactsModel_AddName_EnterAddsContact(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddName
	m.inputNpub = "npub1abc"
	m.inputName = "Alice"
	m2, _ := m.update(pressKey(tea.KeyEnter))
	if len(m2.contacts) != 1 {
		t.Fatalf("contacts count = %d, want 1", len(m2.contacts))
	}
	if m2.contacts[0].PubKey != "npub1abc" {
		t.Errorf("PubKey = %q, want npub1abc", m2.contacts[0].PubKey)
	}
	if m2.contacts[0].DisplayName != "Alice" {
		t.Errorf("DisplayName = %q, want Alice", m2.contacts[0].DisplayName)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
	if m2.state != contactsStateList {
		t.Errorf("expected contactsStateList, got %v", m2.state)
	}
}

func TestContactsModel_AddName_EscCancelsWholeAdd(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddName
	m.inputNpub = "npub1abc"
	m.inputName = "Al"
	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != contactsStateList {
		t.Errorf("expected contactsStateList, got %v", m2.state)
	}
	if m2.inputNpub != "" {
		t.Errorf("inputNpub should be cleared, got %q", m2.inputNpub)
	}
	if m2.inputName != "" {
		t.Errorf("inputName should be cleared, got %q", m2.inputName)
	}
	if len(m2.contacts) != 0 {
		t.Errorf("no contact should be added, got %d", len(m2.contacts))
	}
}

func TestContactsModel_View_EmptyList(t *testing.T) {
	m := makeContactsModel()
	v := m.view(80, 24)
	if !strings.Contains(v, "no contacts") {
		t.Errorf("view should show no-contacts message, got:\n%s", v)
	}
}

func TestContactsModel_View_ShowsCursor(t *testing.T) {
	m := makeContactsModel(
		contactEntry{PubKey: "aaa", DisplayName: "Alice"},
		contactEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	v := m.view(80, 24)
	if !strings.Contains(v, "\x1b[1m") {
		t.Error("view should highlight selected row (bold)")
	}
}

func TestContactsModel_View_ShowsAddNpubPrompt(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	v := m.view(80, 24)
	if !strings.Contains(v, "npub") {
		t.Errorf("view should show npub prompt, got:\n%s", v)
	}
}

func TestContactsModel_View_ShowsAddNamePrompt(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddName
	v := m.view(80, 24)
	if !strings.Contains(v, "display name") {
		t.Errorf("view should show name prompt, got:\n%s", v)
	}
}

func TestContactsModel_AddNpub_Paste(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	m2, _ := m.update(tea.PasteMsg{Content: "npub1abc"})
	if m2.inputNpub != "npub1abc" {
		t.Errorf("inputNpub = %q, want npub1abc", m2.inputNpub)
	}
}

func TestContactsModel_AddName_Paste(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddName
	m2, _ := m.update(tea.PasteMsg{Content: "Alice"})
	if m2.inputName != "Alice" {
		t.Errorf("inputName = %q, want Alice", m2.inputName)
	}
}

func TestContactsModel_Rename_REntersRenameMode(t *testing.T) {
	m := makeContactsModel(contactEntry{PubKey: "aaa", DisplayName: "Alice"})
	m2, _ := m.update(pressChar("r"))
	if m2.state != contactsStateRename {
		t.Errorf("expected contactsStateRename, got %v", m2.state)
	}
	if m2.inputName != "Alice" {
		t.Errorf("inputName = %q, want pre-filled \"Alice\"", m2.inputName)
	}
}

func TestContactsModel_Rename_EnterSavesName(t *testing.T) {
	m := makeContactsModel(contactEntry{PubKey: "aaa", DisplayName: "Alice"})
	m.state = contactsStateRename
	m.inputName = "Alicia"
	m2, _ := m.update(pressKey(tea.KeyEnter))
	if m2.state != contactsStateList {
		t.Errorf("expected contactsStateList, got %v", m2.state)
	}
	if m2.contacts[0].DisplayName != "Alicia" {
		t.Errorf("DisplayName = %q, want \"Alicia\"", m2.contacts[0].DisplayName)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
}

func TestContactsModel_Rename_EscCancels(t *testing.T) {
	m := makeContactsModel(contactEntry{PubKey: "aaa", DisplayName: "Alice"})
	m.state = contactsStateRename
	m.inputName = "Alicia"
	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != contactsStateList {
		t.Errorf("expected contactsStateList, got %v", m2.state)
	}
	if m2.contacts[0].DisplayName != "Alice" {
		t.Errorf("DisplayName should be unchanged, got %q", m2.contacts[0].DisplayName)
	}
}

func TestContactsModel_View_RenameMode_ShowsPrompt(t *testing.T) {
	m := makeContactsModel(contactEntry{PubKey: "aaa", DisplayName: "Alice"})
	m.state = contactsStateRename
	m.inputName = "Alice"
	v := m.view(80, 24)
	if !strings.Contains(v, "Rename contact:") {
		t.Errorf("view should show rename prompt, got:\n%s", v)
	}
	if !strings.Contains(v, "Alice") {
		t.Errorf("view should show current input, got:\n%s", v)
	}
}

func TestContactsModel_View_ShowsError(t *testing.T) {
	m := makeContactsModel()
	m.state = contactsStateAddNpub
	m.err = "npub cannot be empty"
	v := m.view(80, 24)
	if !strings.Contains(v, "npub cannot be empty") {
		t.Errorf("view should show error, got:\n%s", v)
	}
}
