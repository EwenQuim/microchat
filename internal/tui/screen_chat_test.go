package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
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

func TestPubkeyColorDifferentKeysProduceDifferentColors(t *testing.T) {
	r1, g1, b1 := pubkeyColor("03d884d6abcd1234")
	r2, g2, b2 := pubkeyColor("026df5f7efgh5678")

	if r1 == r2 && g1 == g2 && b1 == b2 {
		t.Errorf("different pubkeys should produce different colors, both got (%d,%d,%d)", r1, g1, b1)
	}
}

func TestPubkeyColorConsistency(t *testing.T) {
	r1, g1, b1 := pubkeyColor("03d884d6abcd1234")
	r2, g2, b2 := pubkeyColor("03d884d6abcd1234")

	if r1 != r2 || g1 != g2 || b1 != b2 {
		t.Errorf("same pubkey should produce same color")
	}
}

func TestChatModel_View_PubkeyOnlyDisplayName(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	pk := id.PubKeyHex
	content := "hello world"
	msg := generated.Message{
		Pubkey:  &pk,
		Content: &content,
	}
	m := newChatModel(nil, "room", "", nil, "alice")
	m.loading = false
	m.messages = []generated.Message{msg}

	v := m.viewPanel(60, 10, true)

	// truncPk is last 6 chars of npub
	npub := id.NpubKey
	truncPk := npub[len(npub)-6:]
	if !strings.Contains(v, truncPk) {
		t.Errorf("view should contain %s for pubkey-only message, got:\n%s", truncPk, v)
	}
	if !strings.Contains(v, truncPk+"…") {
		t.Errorf("view should show truncated npub %s… as display name, got:\n%s", truncPk, v)
	}
}

func BenchmarkViewPanel_ColorCache(b *testing.B) {
	pk := "03d884d6abcdef1234567890"
	content := "hello world"
	msgs := make([]generated.Message, 40)
	for i := range msgs {
		p := pk
		c := content
		msgs[i] = generated.Message{Pubkey: &p, Content: &c}
	}
	m := newChatModel(nil, "room", "", nil, "alice")
	m.loading = false
	m.messages = msgs

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.viewPanel(80, 45, true)
	}
}

// Verify hue is spread across spectrum, not all red
func TestPubkeyColorNotAllRed(t *testing.T) {
	keys := []string{
		"03d884d6", "026df5f7", "037e530d", "03dffe83", "03a4ad4c",
	}
	allSameColor := true
	r0, g0, b0 := pubkeyColor(keys[0])
	for _, k := range keys[1:] {
		r, g, b := pubkeyColor(k)
		if r != r0 || g != g0 || b != b0 {
			allSameColor = false
			break
		}
	}
	if allSameColor {
		t.Errorf("all pubkeys mapped to same color (%d,%d,%d) — hue scaling bug", r0, g0, b0)
	}
}
