package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// pressRealChar simulates a real terminal keypress where both Code and Text are set.
func pressRealChar(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text}
}

func TestChatModel_TypingMode_SpaceAppended(t *testing.T) {
	m := newChatModel(nil, "room", "", nil, "alice")
	m.typing = true
	m.inputText = "hello"

	m2, _ := m.update(pressRealChar(' ', " "))

	if m2.inputText != "hello " {
		t.Errorf("inputText = %q, want %q (space not appended)", m2.inputText, "hello ")
	}
}

func TestChatModel_TypingMode_RAppended(t *testing.T) {
	m := newChatModel(nil, "room", "", nil, "alice")
	m.typing = true
	m.inputText = "hello"

	m2, _ := m.update(pressRealChar('r', "r"))

	if m2.inputText != "hellor" {
		t.Errorf("inputText = %q, want %q (r not appended in typing mode)", m2.inputText, "hellor")
	}
}

func TestChatModel_NormalMode_IEntersTyping(t *testing.T) {
	m := newChatModel(nil, "room", "", nil, "alice")
	// typing starts false

	m2, _ := m.update(pressRealChar('i', "i"))

	if !m2.typing {
		t.Error("expected typing=true after pressing i in normal mode")
	}
}

func TestChatModel_NormalMode_RRefreshes(t *testing.T) {
	m := newChatModel(nil, "room", "", nil, "alice")
	m.loading = false
	// typing starts false

	m2, cmd := m.update(pressRealChar('r', "r"))

	if !m2.loading {
		t.Error("expected loading=true after pressing r in normal mode")
	}
	if cmd == nil {
		t.Error("expected fetch cmd after pressing r in normal mode")
	}
}

func TestChatModel_TypingMode_EscExitsTyping(t *testing.T) {
	m := newChatModel(nil, "room", "", nil, "alice")
	m.typing = true

	m2, _ := m.update(pressKey(tea.KeyEscape))

	if m2.typing {
		t.Error("expected typing=false after pressing Esc")
	}
}

func TestChatModel_View_TypingModeCursor(t *testing.T) {
	m := newChatModel(nil, "room", "", nil, "alice")
	m.typing = true
	m.loading = false

	v := m.viewPanel(60, 10, true)
	if !strings.Contains(v, "█") {
		t.Error("view in typing mode should show cursor █")
	}
}

func TestChatModel_View_NormalModeNoCursor(t *testing.T) {
	m := newChatModel(nil, "room", "", nil, "alice")
	m.typing = false
	m.loading = false

	v := m.viewPanel(60, 10, true)
	if strings.Contains(v, "█") {
		t.Error("view in normal mode should not show cursor █")
	}
}
