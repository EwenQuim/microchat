package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func makeUsersModel(users ...userEntry) usersModel {
	return newUsersModel(appConfig{Users: users})
}

func TestUsersModel_NewFromConfig(t *testing.T) {
	m := makeUsersModel(userEntry{PubKey: "abc123", DisplayName: "Alice"})
	if len(m.users) != 1 {
		t.Errorf("users count = %d, want 1", len(m.users))
	}
	if m.users[0].PubKey != "abc123" {
		t.Errorf("PubKey = %q, want abc123", m.users[0].PubKey)
	}
}

func TestUsersModel_CursorUp(t *testing.T) {
	m := makeUsersModel(
		userEntry{PubKey: "aaa", DisplayName: "Alice"},
		userEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m.cursor = 1
	m2, _ := m.update(pressKey(tea.KeyUp))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m2.cursor)
	}
}

func TestUsersModel_CursorUp_k(t *testing.T) {
	m := makeUsersModel(
		userEntry{PubKey: "aaa", DisplayName: "Alice"},
		userEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m.cursor = 1
	m2, _ := m.update(pressChar("k"))
	if m2.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", m2.cursor)
	}
}

func TestUsersModel_CursorUp_Bounded(t *testing.T) {
	m := makeUsersModel(userEntry{PubKey: "aaa"})
	m2, _ := m.update(pressKey(tea.KeyUp))
	if m2.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", m2.cursor)
	}
}

func TestUsersModel_CursorDown(t *testing.T) {
	m := makeUsersModel(
		userEntry{PubKey: "aaa", DisplayName: "Alice"},
		userEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m2.cursor)
	}
}

func TestUsersModel_CursorDown_j(t *testing.T) {
	m := makeUsersModel(
		userEntry{PubKey: "aaa", DisplayName: "Alice"},
		userEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m2, _ := m.update(pressChar("j"))
	if m2.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m2.cursor)
	}
}

func TestUsersModel_CursorDown_Bounded(t *testing.T) {
	m := makeUsersModel(userEntry{PubKey: "aaa"})
	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 0 {
		t.Errorf("cursor should not exceed len-1, got %d", m2.cursor)
	}
}

func TestUsersModel_Delete(t *testing.T) {
	m := makeUsersModel(
		userEntry{PubKey: "aaa", DisplayName: "Alice"},
		userEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	m.cursor = 0
	m2, _ := m.update(pressChar("d"))
	if len(m2.users) != 1 {
		t.Errorf("users count = %d, want 1", len(m2.users))
	}
	if m2.users[0].PubKey != "bbb" {
		t.Errorf("remaining user = %q, want bbb", m2.users[0].PubKey)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
}

func TestUsersModel_Delete_AdjustsCursor(t *testing.T) {
	m := makeUsersModel(
		userEntry{PubKey: "aaa"},
		userEntry{PubKey: "bbb"},
	)
	m.cursor = 1
	m2, _ := m.update(pressChar("d"))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after deleting last item", m2.cursor)
	}
}

func TestUsersModel_Delete_EmptyList(t *testing.T) {
	m := makeUsersModel()
	m2, _ := m.update(pressChar("d"))
	if len(m2.users) != 0 {
		t.Errorf("users count = %d, want 0", len(m2.users))
	}
}

func TestUsersModel_Esc_NavigatesToServers(t *testing.T) {
	m := makeUsersModel()
	_, cmd := m.update(pressKey(tea.KeyEscape))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenServers {
		t.Errorf("expected screenServers, got %v", nav.to)
	}
}

func TestUsersModel_Tab_NavigatesToServers(t *testing.T) {
	m := makeUsersModel()
	_, cmd := m.update(pressKey(tea.KeyTab))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenServers {
		t.Errorf("expected screenServers, got %v", nav.to)
	}
}

func TestUsersModel_PressA_EntersAddNpubState(t *testing.T) {
	m := makeUsersModel()
	m2, _ := m.update(pressChar("a"))
	if m2.state != usersStateAddNpub {
		t.Errorf("expected usersStateAddNpub, got %v", m2.state)
	}
}

func TestUsersModel_AddNpub_TypingBuildsInput(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddNpub
	m2, _ := m.update(pressChar("n"))
	if m2.inputNpub != "n" {
		t.Errorf("inputNpub = %q, want \"n\"", m2.inputNpub)
	}
}

func TestUsersModel_AddNpub_Backspace(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddNpub
	m.inputNpub = "ab"
	m2, _ := m.update(pressKey(tea.KeyBackspace))
	if m2.inputNpub != "a" {
		t.Errorf("inputNpub = %q, want \"a\"", m2.inputNpub)
	}
}

func TestUsersModel_AddNpub_EscCancels(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddNpub
	m.inputNpub = "abc"
	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != usersStateList {
		t.Errorf("expected usersStateList, got %v", m2.state)
	}
	if m2.inputNpub != "" {
		t.Errorf("inputNpub should be cleared, got %q", m2.inputNpub)
	}
}

func TestUsersModel_AddNpub_EnterEmpty_Error(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddNpub
	m2, cmd := m.update(pressKey(tea.KeyEnter))
	if m2.err == "" {
		t.Error("expected error for empty npub")
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty npub")
	}
}

func TestUsersModel_AddNpub_EnterTransitionsToAddName(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddNpub
	m.inputNpub = "npub1somekey"
	m2, _ := m.update(pressKey(tea.KeyEnter))
	if m2.state != usersStateAddName {
		t.Errorf("expected usersStateAddName, got %v", m2.state)
	}
}

func TestUsersModel_AddName_TypingBuildsInput(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddName
	m.inputNpub = "npub1key"
	m2, _ := m.update(pressChar("A"))
	if m2.inputName != "A" {
		t.Errorf("inputName = %q, want \"A\"", m2.inputName)
	}
}

func TestUsersModel_AddName_Backspace(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddName
	m.inputName = "Al"
	m2, _ := m.update(pressKey(tea.KeyBackspace))
	if m2.inputName != "A" {
		t.Errorf("inputName = %q, want \"A\"", m2.inputName)
	}
}

func TestUsersModel_AddName_EnterAddsUser(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddName
	m.inputNpub = "npub1abc"
	m.inputName = "Alice"
	m2, _ := m.update(pressKey(tea.KeyEnter))
	if len(m2.users) != 1 {
		t.Fatalf("users count = %d, want 1", len(m2.users))
	}
	if m2.users[0].PubKey != "npub1abc" {
		t.Errorf("PubKey = %q, want npub1abc", m2.users[0].PubKey)
	}
	if m2.users[0].DisplayName != "Alice" {
		t.Errorf("DisplayName = %q, want Alice", m2.users[0].DisplayName)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
	if m2.state != usersStateList {
		t.Errorf("expected usersStateList, got %v", m2.state)
	}
}

func TestUsersModel_AddName_EscCancelsWholeAdd(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddName
	m.inputNpub = "npub1abc"
	m.inputName = "Al"
	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != usersStateList {
		t.Errorf("expected usersStateList, got %v", m2.state)
	}
	if m2.inputNpub != "" {
		t.Errorf("inputNpub should be cleared, got %q", m2.inputNpub)
	}
	if m2.inputName != "" {
		t.Errorf("inputName should be cleared, got %q", m2.inputName)
	}
	if len(m2.users) != 0 {
		t.Errorf("no user should be added, got %d", len(m2.users))
	}
}

func TestUsersModel_View_EmptyList(t *testing.T) {
	m := makeUsersModel()
	v := m.view(80, 24)
	if !strings.Contains(v, "no contacts") {
		t.Errorf("view should show no-contacts message, got:\n%s", v)
	}
}

func TestUsersModel_View_ShowsCursor(t *testing.T) {
	m := makeUsersModel(
		userEntry{PubKey: "aaa", DisplayName: "Alice"},
		userEntry{PubKey: "bbb", DisplayName: "Bob"},
	)
	v := m.view(80, 24)
	if !strings.Contains(v, "▶") {
		t.Error("view should show cursor ▶")
	}
}

func TestUsersModel_View_ShowsAddNpubPrompt(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddNpub
	v := m.view(80, 24)
	if !strings.Contains(v, "npub") {
		t.Errorf("view should show npub prompt, got:\n%s", v)
	}
}

func TestUsersModel_View_ShowsAddNamePrompt(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddName
	v := m.view(80, 24)
	if !strings.Contains(v, "display name") {
		t.Errorf("view should show name prompt, got:\n%s", v)
	}
}

func TestUsersModel_View_ShowsError(t *testing.T) {
	m := makeUsersModel()
	m.state = usersStateAddNpub
	m.err = "npub cannot be empty"
	v := m.view(80, 24)
	if !strings.Contains(v, "npub cannot be empty") {
		t.Errorf("view should show error, got:\n%s", v)
	}
}
