package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// runCmd executes a Cmd and returns its Msg.
func runCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// pressChar simulates pressing a single printable character.
func pressChar(ch string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Text: ch}
}

// pressKey simulates pressing a special key by its code.
func pressKey(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func TestIdentityModel_MenuPress_g(t *testing.T) {
	m := newIdentityModel()
	m2, cmd := m.update(pressChar("g"))

	if m2.result.privKey == nil {
		t.Error("expected result identity to be set after pressing g")
	}
	if m2.err != "" {
		t.Errorf("unexpected error: %s", m2.err)
	}

	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenServers {
		t.Errorf("expected navigate to screenServers, got %v", nav.to)
	}
}

func TestIdentityModel_MenuPress_p(t *testing.T) {
	m := newIdentityModel()
	m2, cmd := m.update(pressChar("p"))

	if m2.state != idStateInput {
		t.Errorf("expected idStateInput, got %v", m2.state)
	}
	if cmd != nil {
		t.Error("expected nil cmd after pressing p")
	}
}

func TestIdentityModel_MenuPress_q(t *testing.T) {
	m := newIdentityModel()
	_, cmd := m.update(pressChar("q"))

	if cmd == nil {
		t.Fatal("expected quit cmd")
	}
	msg := runCmd(cmd)
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestIdentityModel_InputState_TypingAppends(t *testing.T) {
	m := newIdentityModel()
	m.state = idStateInput
	m.inputText = "ab"

	m2, _ := m.update(pressChar("c"))
	if m2.inputText != "abc" {
		t.Errorf("inputText = %q, want %q", m2.inputText, "abc")
	}
}

func TestIdentityModel_InputState_Backspace(t *testing.T) {
	m := newIdentityModel()
	m.state = idStateInput
	m.inputText = "abc"

	m2, _ := m.update(pressKey(tea.KeyBackspace))
	if m2.inputText != "ab" {
		t.Errorf("inputText = %q, want %q", m2.inputText, "ab")
	}
}

func TestIdentityModel_InputState_BackspaceEmpty(t *testing.T) {
	m := newIdentityModel()
	m.state = idStateInput

	m2, _ := m.update(pressKey(tea.KeyBackspace))
	if m2.inputText != "" {
		t.Errorf("inputText should remain empty, got %q", m2.inputText)
	}
}

func TestIdentityModel_InputState_Esc(t *testing.T) {
	m := newIdentityModel()
	m.state = idStateInput
	m.inputText = "some text"
	m.err = "previous error"

	m2, cmd := m.update(pressKey(tea.KeyEscape))
	if m2.state != idStateMenu {
		t.Errorf("expected idStateMenu, got %v", m2.state)
	}
	if m2.inputText != "" {
		t.Errorf("inputText should be cleared, got %q", m2.inputText)
	}
	if m2.err != "" {
		t.Errorf("err should be cleared, got %q", m2.err)
	}
	if cmd != nil {
		t.Error("expected nil cmd after esc")
	}
}

func TestIdentityModel_InputState_EnterValidKey(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}

	m := newIdentityModel()
	m.state = idStateInput
	m.inputText = id.PrivKeyHex

	m2, cmd := m.update(pressKey(tea.KeyEnter))
	if m2.err != "" {
		t.Errorf("unexpected error: %s", m2.err)
	}
	if m2.result.privKey == nil {
		t.Error("expected result identity to be set")
	}

	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenServers {
		t.Errorf("expected screenServers, got %v", nav.to)
	}
}

func TestIdentityModel_InputState_EnterInvalidKey(t *testing.T) {
	m := newIdentityModel()
	m.state = idStateInput
	m.inputText = "notvalidhex!!!"

	m2, cmd := m.update(pressKey(tea.KeyEnter))
	if m2.err == "" {
		t.Error("expected error for invalid key")
	}
	if m2.state != idStateMenu {
		t.Errorf("expected idStateMenu after error, got %v", m2.state)
	}
	if cmd != nil {
		t.Error("expected nil cmd on error")
	}
}

func TestIdentityModel_View_MenuContainsOptions(t *testing.T) {
	m := newIdentityModel()
	v := m.view(80, 24)

	if !strings.Contains(v, "[g]") {
		t.Error("view should contain [g]")
	}
	if !strings.Contains(v, "[p]") {
		t.Error("view should contain [p]")
	}
	if !strings.Contains(v, "[q]") {
		t.Error("view should contain [q]")
	}
}

func TestIdentityModel_View_InputShowsCursor(t *testing.T) {
	m := newIdentityModel()
	m.state = idStateInput

	v := m.view(80, 24)
	if !strings.Contains(v, "█") {
		t.Error("view in input state should show cursor █")
	}
}

func TestIdentityModel_View_ShowsError(t *testing.T) {
	m := newIdentityModel()
	m.err = "something went wrong"

	v := m.view(80, 24)
	if !strings.Contains(v, "something went wrong") {
		t.Error("view should display the error message")
	}
}
