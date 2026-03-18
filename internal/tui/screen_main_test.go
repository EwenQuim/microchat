package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func makeMainModelWithChat() mainModel {
	m := newMainModel(nil, nil, nil, "", nil)
	m.chat = chatModel{room: "general"}
	m.hasChat = true
	return m
}

func sendMainKey(m mainModel, code rune) mainModel {
	updated, _ := m.update(tea.KeyPressMsg{Code: code})
	return updated
}

// TestMainModel_RightArrow_ShiftsFocusToChat verifies → moves focus to chat when hasChat.
func TestMainModel_RightArrow_ShiftsFocusToChat(t *testing.T) {
	m := makeMainModelWithChat()
	m.focus = focusLeft

	m = sendMainKey(m, tea.KeyRight)

	if m.focus != focusRight {
		t.Errorf("focus = %v, want focusRight after → with chat loaded", m.focus)
	}
}

// TestMainModel_RightArrow_NoChat_NoOp verifies → does nothing when no chat is loaded.
func TestMainModel_RightArrow_NoChat_NoOp(t *testing.T) {
	m := newMainModel(nil, nil, nil, "", nil)
	m.focus = focusLeft

	m = sendMainKey(m, tea.KeyRight)

	if m.focus != focusLeft {
		t.Errorf("focus = %v, want focusLeft (no-op) after → with no chat", m.focus)
	}
}

// TestMainModel_LeftArrow_ShiftsFocusToRooms verifies ← moves focus to rooms when on right.
func TestMainModel_LeftArrow_ShiftsFocusToRooms(t *testing.T) {
	m := makeMainModelWithChat()
	m.focus = focusRight

	m = sendMainKey(m, tea.KeyLeft)

	if m.focus != focusLeft {
		t.Errorf("focus = %v, want focusLeft after ← from right panel", m.focus)
	}
}

// TestMainModel_LeftArrow_AlreadyLeft_NoOp verifies ← does nothing when already on left.
func TestMainModel_LeftArrow_AlreadyLeft_NoOp(t *testing.T) {
	m := newMainModel(nil, nil, nil, "", nil)
	m.focus = focusLeft

	m = sendMainKey(m, tea.KeyLeft)

	if m.focus != focusLeft {
		t.Errorf("focus = %v, want focusLeft (no-op) after ← already on left", m.focus)
	}
}
