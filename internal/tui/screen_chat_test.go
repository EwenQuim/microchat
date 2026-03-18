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
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.typing = true
	m.inputText = "hello"

	m2, _ := m.update(pressRealChar(' ', " "))

	if m2.inputText != "hello " {
		t.Errorf("inputText = %q, want %q (space not appended)", m2.inputText, "hello ")
	}
}

func TestChatModel_TypingMode_RAppended(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.typing = true
	m.inputText = "hello"

	m2, _ := m.update(pressRealChar('r', "r"))

	if m2.inputText != "hellor" {
		t.Errorf("inputText = %q, want %q (r not appended in typing mode)", m2.inputText, "hellor")
	}
}

func TestChatModel_NormalMode_IEntersTyping(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	// typing starts false

	m2, _ := m.update(pressRealChar('i', "i"))

	if !m2.typing {
		t.Error("expected typing=true after pressing i in normal mode")
	}
}

func TestChatModel_NormalMode_RRefreshes(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
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
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.typing = true

	m2, _ := m.update(pressKey(tea.KeyEscape))

	if m2.typing {
		t.Error("expected typing=false after pressing Esc")
	}
}

func TestChatModel_View_TypingModeCursor(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.typing = true
	m.loading = false

	v := m.viewPanel(60, 10, true)
	if !strings.Contains(v, "█") {
		t.Error("view in typing mode should show cursor █")
	}
}

func TestChatModel_View_NormalModeNoCursor(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
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
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
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
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.loading = false
	m.messages = msgs

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.viewPanel(80, 45, true)
	}
}

func TestChatModel_CursorMode_VEnters(t *testing.T) {
	pk := "abc123"
	content := "hello"
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.loading = false
	m.messages = []generated.Message{
		{Pubkey: &pk, Content: &content},
	}

	m2, _ := m.update(pressRealChar('v', "v"))

	if !m2.msgCursorMode {
		t.Error("expected msgCursorMode=true after pressing v")
	}
	if m2.msgCursor != 0 {
		t.Errorf("msgCursor = %d, want 0 (last message)", m2.msgCursor)
	}
}

func TestChatModel_CursorMode_EscExits(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.loading = false
	m.msgCursorMode = true
	m.msgCursor = 0

	m2, _ := m.update(pressKey(tea.KeyEscape))

	if m2.msgCursorMode {
		t.Error("expected msgCursorMode=false after pressing Esc in cursor mode")
	}
}

func TestChatModel_CursorMode_UpMovesUp(t *testing.T) {
	pk := "abc"
	content := "msg"
	msgs := []generated.Message{
		{Pubkey: &pk, Content: &content},
		{Pubkey: &pk, Content: &content},
		{Pubkey: &pk, Content: &content},
	}
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.loading = false
	m.messages = msgs
	m.msgCursorMode = true
	m.msgCursor = 2 // start at last

	m2, _ := m.update(pressKey(tea.KeyUp))

	if m2.msgCursor != 1 {
		t.Errorf("msgCursor = %d, want 1 after pressing up", m2.msgCursor)
	}
}

func TestChatModel_CursorMode_AEntersRenameMode(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	pk := id.PubKeyHex
	user := "alice"
	content := "hello"
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "bob")
	m.loading = false
	m.messages = []generated.Message{
		{Pubkey: &pk, User: &user, Content: &content},
	}
	m.msgCursorMode = true
	m.msgCursor = 0

	m2, cmd := m.update(pressRealChar('a', "a"))

	if cmd != nil {
		t.Error("expected nil cmd — rename mode should not emit immediately")
	}
	if !m2.chatRenameMode {
		t.Error("expected chatRenameMode=true after pressing a")
	}
	if m2.renameInput != user {
		t.Errorf("renameInput = %q, want %q (pre-filled from message user)", m2.renameInput, user)
	}
	if m2.pendingPubKey != pk {
		t.Errorf("pendingPubKey = %q, want %q", m2.pendingPubKey, pk)
	}
}

func TestChatModel_RenameMode_TypingAppendsToInput(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "bob")
	m.chatRenameMode = true
	m.renameInput = "alice"

	m2, _ := m.update(pressRealChar('x', "x"))

	if m2.renameInput != "alicex" {
		t.Errorf("renameInput = %q, want %q", m2.renameInput, "alicex")
	}
}

func TestChatModel_RenameMode_BackspaceDeletes(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "bob")
	m.chatRenameMode = true
	m.renameInput = "alice"

	m2, _ := m.update(pressKey(tea.KeyBackspace))

	if m2.renameInput != "alic" {
		t.Errorf("renameInput = %q, want %q after backspace", m2.renameInput, "alic")
	}
}

func TestChatModel_RenameMode_EscExits(t *testing.T) {
	pk := "abc123"
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "bob")
	m.chatRenameMode = true
	m.renameInput = "alice"
	m.pendingPubKey = pk

	m2, _ := m.update(pressKey(tea.KeyEscape))

	if m2.chatRenameMode {
		t.Error("expected chatRenameMode=false after esc")
	}
	if m2.renameInput != "" {
		t.Errorf("renameInput = %q, want empty after esc", m2.renameInput)
	}
	if m2.pendingPubKey != "" {
		t.Errorf("pendingPubKey = %q, want empty after esc", m2.pendingPubKey)
	}
}

func TestChatModel_RenameMode_EnterEmitsMsg(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	pk := id.PubKeyHex
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "bob")
	m.chatRenameMode = true
	m.renameInput = "alice renamed"
	m.pendingPubKey = pk

	m2, cmd := m.update(pressKey(tea.KeyEnter))

	if cmd == nil {
		t.Fatal("expected non-nil cmd after pressing enter in rename mode")
	}
	result := cmd()
	ac, ok := result.(addContactFromChatMsg)
	if !ok {
		t.Fatalf("expected addContactFromChatMsg, got %T", result)
	}
	if ac.pubKeyHex != pk {
		t.Errorf("pubKeyHex = %q, want %q", ac.pubKeyHex, pk)
	}
	if ac.displayName != "alice renamed" {
		t.Errorf("displayName = %q, want %q", ac.displayName, "alice renamed")
	}
	if m2.chatRenameMode {
		t.Error("expected chatRenameMode=false after enter")
	}
	if !strings.Contains(m2.statusMsg, "alice renamed") {
		t.Errorf("statusMsg = %q, expected to contain contact name", m2.statusMsg)
	}
}

func TestChatModel_View_RenameMode_ShowsPrompt(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "bob")
	m.loading = false
	m.chatRenameMode = true
	m.renameInput = "alice"

	v := m.viewPanel(60, 10, true)

	if !strings.Contains(v, "Add contact as:") {
		t.Errorf("expected 'Add contact as:' in view when chatRenameMode, got:\n%s", v)
	}
	if !strings.Contains(v, "alice") {
		t.Errorf("expected renameInput 'alice' in view, got:\n%s", v)
	}
}

func TestChatModel_CursorMode_AWithNoPubkey_ShowsError(t *testing.T) {
	content := "hello"
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "bob")
	m.loading = false
	m.messages = []generated.Message{
		{Content: &content}, // no pubkey
	}
	m.msgCursorMode = true
	m.msgCursor = 0

	m2, cmd := m.update(pressRealChar('a', "a"))

	if m2.err == "" {
		t.Error("expected error when pressing a on message with no pubkey")
	}
	if cmd != nil {
		t.Error("expected nil cmd when no pubkey")
	}
}

func TestChatModel_View_CursorMode_ShowsArrow(t *testing.T) {
	pk := "abc123"
	content := "hello"
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.loading = false
	m.messages = []generated.Message{
		{Pubkey: &pk, Content: &content},
	}
	m.msgCursorMode = true
	m.msgCursor = 0

	v := m.viewPanel(60, 10, true)

	if !strings.Contains(v, "▶") {
		t.Errorf("expected ▶ in view when in cursor mode, got:\n%s", v)
	}
}

func TestChatModel_View_ContactShowsDisplayName(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	pk := id.PubKeyHex
	user := "original_username"
	content := "hello world"
	msg := generated.Message{
		Pubkey:  &pk,
		User:    &user,
		Content: &content,
	}
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "me")
	m.loading = false
	m.messages = []generated.Message{msg}
	m.contacts = []contactEntry{{PubKey: pk, DisplayName: "Alice Contact"}}

	v := m.viewPanel(80, 10, true)

	if !strings.Contains(v, "Alice Contact") {
		t.Errorf("view should show contact display name, got:\n%s", v)
	}
	if strings.Contains(v, user) {
		t.Errorf("view should not show original username %q when contact has display name, got:\n%s", user, v)
	}
}

func TestChatModel_View_ContactHidesKeyLabel(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	pk := id.PubKeyHex
	user := "original_username"
	content := "hello world"
	msg := generated.Message{
		Pubkey:  &pk,
		User:    &user,
		Content: &content,
	}
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "me")
	m.loading = false
	m.messages = []generated.Message{msg}
	m.contacts = []contactEntry{{PubKey: pk, DisplayName: "Alice Contact"}}

	v := m.viewPanel(80, 10, true)

	npub := id.NpubKey
	truncPk := npub[len(npub)-8:]
	if strings.Contains(v, truncPk) {
		t.Errorf("view should hide truncated npub for known contact, got:\n%s", v)
	}
}

func TestChatModel_View_ContactCheckmark(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	pk := id.PubKeyHex
	user := "original_username"
	content := "hello world"
	msg := generated.Message{
		Pubkey:  &pk,
		User:    &user,
		Content: &content,
	}
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "me")
	m.loading = false
	m.messages = []generated.Message{msg}
	m.contacts = []contactEntry{{PubKey: pk, DisplayName: "Alice Contact"}}

	v := m.viewPanel(80, 10, true)

	if !strings.Contains(v, "✓") {
		t.Errorf("view should show ✓ for known contact, got:\n%s", v)
	}
}

func TestChatModel_View_NonContactShowsUser(t *testing.T) {
	id, err := generateIdentity()
	if err != nil {
		t.Fatalf("generateIdentity: %v", err)
	}
	pk := id.PubKeyHex
	user := "original_username"
	content := "hello world"
	msg := generated.Message{
		Pubkey:  &pk,
		User:    &user,
		Content: &content,
	}
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "me")
	m.loading = false
	m.messages = []generated.Message{msg}
	// No contacts set

	v := m.viewPanel(80, 10, true)

	if !strings.Contains(v, user) {
		t.Errorf("view should show original username for non-contact, got:\n%s", v)
	}
	if strings.Contains(v, "✓") {
		t.Errorf("view should not show ✓ for non-contact, got:\n%s", v)
	}
}

func TestChatModel_SendWithNoIdentity_ReturnsError(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	cmd := m.sendMessage("hello")
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	result := cmd()
	msg, ok := result.(messageSentMsg)
	if !ok {
		t.Fatalf("expected messageSentMsg, got %T", result)
	}
	if msg.err == nil {
		t.Fatal("expected error when no identity configured")
	}
	if !strings.Contains(msg.err.Error(), "no identity") {
		t.Errorf("error = %q, want it to contain 'no identity'", msg.err.Error())
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
