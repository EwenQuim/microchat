package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func makeMainModelWithChat() mainModel {
	m := newMainModel(appConfig{}, nil, nil, nil, "", nil)
	m.chat = chatModel{room: "general"}
	m.hasChat = true
	m.right = rightChat
	return m
}

func makeMainModelWithServers() mainModel {
	cfg := appConfig{Servers: []serverConfig{{URL: "http://alpha.example", Quickname: "Alpha"}}}
	return newMainModel(cfg, nil, cfg.Servers, nil, "", nil)
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
	m := newMainModel(appConfig{}, nil, nil, nil, "", nil)
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
	m := newMainModel(appConfig{}, nil, nil, nil, "", nil)
	m.focus = focusLeft

	m = sendMainKey(m, tea.KeyLeft)

	if m.focus != focusLeft {
		t.Errorf("focus = %v, want focusLeft (no-op) after ← already on left", m.focus)
	}
}

// TestMainModel_SectionSelected_SetsRightAndFocus verifies a focused sectionSelectedMsg
// switches the right pane and moves focus; an unfocused one previews without focusing.
func TestMainModel_SectionSelected_SetsRightAndFocus(t *testing.T) {
	m := makeMainModelWithServers()

	m2, _ := m.update(sectionSelectedMsg{to: screenServers, focus: true})
	if m2.right != rightServers {
		t.Errorf("right = %v, want rightServers", m2.right)
	}
	if m2.focus != focusRight {
		t.Errorf("focus = %v, want focusRight after focused section select", m2.focus)
	}

	m3, _ := m.update(sectionSelectedMsg{to: screenContacts, focus: false})
	if m3.right != rightContacts {
		t.Errorf("right = %v, want rightContacts", m3.right)
	}
	if m3.focus != focusLeft {
		t.Errorf("focus = %v, want focusLeft after preview select", m3.focus)
	}
}

// TestMainModel_ShortcutS_PreviewsServers verifies the S shortcut moves the sidebar cursor
// to the Servers nav item and emits an unfocused section preview.
func TestMainModel_ShortcutS_PreviewsServers(t *testing.T) {
	m := makeMainModelWithServers()

	m2, cmd := m.update(pressChar("S"))
	if got := m2.navIndexFor(screenServers); m2.rooms.cursor != got {
		t.Errorf("rooms.cursor = %d, want %d (Servers nav index)", m2.rooms.cursor, got)
	}
	if m2.focus != focusLeft {
		t.Errorf("focus = %v, want focusLeft after S shortcut", m2.focus)
	}
	sec, ok := runCmd(cmd).(sectionSelectedMsg)
	if !ok {
		t.Fatalf("expected sectionSelectedMsg from S shortcut, got %T", runCmd(cmd))
	}
	if sec.to != screenServers || sec.focus {
		t.Errorf("got %+v, want {screenServers, focus:false}", sec)
	}
}

// TestMainModel_Tab_TogglesFocus verifies Tab flips focus between sidebar and right pane.
func TestMainModel_Tab_TogglesFocus(t *testing.T) {
	m := makeMainModelWithChat()
	m.focus = focusLeft

	m = sendMainKey(m, tea.KeyTab)
	if m.focus != focusRight {
		t.Fatalf("focus = %v, want focusRight after first Tab", m.focus)
	}
	m = sendMainKey(m, tea.KeyTab)
	if m.focus != focusLeft {
		t.Errorf("focus = %v, want focusLeft after second Tab", m.focus)
	}
}

// TestMainModel_LeftFromConfig_ReturnsToSidebar verifies ←/Esc exit a focused config pane.
func TestMainModel_LeftFromConfig_ReturnsToSidebar(t *testing.T) {
	for _, key := range []rune{tea.KeyLeft, tea.KeyEscape} {
		m := makeMainModelWithServers()
		m.right = rightServers
		m.focus = focusRight

		m = sendMainKey(m, key)
		if m.focus != focusLeft {
			t.Errorf("key %v: focus = %v, want focusLeft", key, m.focus)
		}
	}
}

// TestMainModel_ConfigFocused_A_Delegates verifies list keys reach the config model.
func TestMainModel_ConfigFocused_A_Delegates(t *testing.T) {
	m := makeMainModelWithServers()
	m.right = rightServers
	m.focus = focusRight

	m2, _ := m.update(pressChar("a"))
	if m2.serversSec.state != serverStateAddURL {
		t.Errorf("serversSec.state = %v, want serverStateAddURL after delegating 'a'", m2.serversSec.state)
	}
}

// TestMainModel_ConfigFocused_ServerEnter_NotDelegated verifies Enter on a focused Servers
// list does not bubble a navigateMsg (which would reset the main model).
func TestMainModel_ConfigFocused_ServerEnter_NotDelegated(t *testing.T) {
	m := makeMainModelWithServers()
	m.right = rightServers
	m.focus = focusRight

	m2, cmd := m.update(pressKey(tea.KeyEnter))
	if cmd != nil {
		if _, ok := runCmd(cmd).(navigateMsg); ok {
			t.Error("Enter on in-pane Servers should not emit navigateMsg")
		}
	}
	if m2.right != rightServers || m2.focus != focusRight {
		t.Errorf("Enter should be a no-op in-pane; got right=%v focus=%v", m2.right, m2.focus)
	}
}

// TestMainModel_View_ConfigPane_NoChatChrome verifies the right pane shows the section
// title and not the chat help/input chrome.
func TestMainModel_View_ConfigPane_NoChatChrome(t *testing.T) {
	m := makeMainModelWithServers()
	m.right = rightServers
	m.focus = focusRight

	v := m.view(80, 12)
	if !strings.Contains(v, "Servers") {
		t.Errorf("config pane should show the Servers title, got:\n%s", v)
	}
	if strings.Contains(v, "insert") {
		t.Errorf("config pane should not show chat help chrome, got:\n%s", v)
	}
}

// TestServerModel_ViewPanel_TitleHelpTruncation verifies the in-pane Servers section renders
// its title + help bar and truncates long URLs to the pane width.
func TestServerModel_ViewPanel_TitleHelpTruncation(t *testing.T) {
	longURL := "https://verylongserverurl.example.com/some/deep/path"
	m := newServerModel(appConfig{Servers: []serverConfig{{URL: longURL, Quickname: "Alpha"}}})

	v := m.viewPanel(40, 12, true)
	if !strings.Contains(v, "Servers") {
		t.Errorf("missing Servers title:\n%s", v)
	}
	if !strings.Contains(v, "Alpha") {
		t.Errorf("missing server name:\n%s", v)
	}
	if !strings.Contains(v, "add") {
		t.Errorf("missing help bar:\n%s", v)
	}
	if strings.Contains(v, longURL) {
		t.Errorf("long URL should be truncated in a 40-wide pane:\n%s", v)
	}
}

// TestConfigViewPanels_NoPanicEmpty verifies each config section renders without panicking
// when empty and small.
func TestConfigViewPanels_NoPanicEmpty(t *testing.T) {
	cfg := appConfig{}
	_ = newServerModel(cfg).viewPanel(20, 4, false)
	_ = newIdentitiesModel(cfg).viewPanel(20, 4, true)
	_ = newContactsModel(cfg).viewPanel(20, 4, true)
}

// TestSyncMainConfig_PersistsContacts verifies in-pane contact edits are saved and propagated.
func TestSyncMainConfig_PersistsContacts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := model{cfg: appConfig{}}
	m.main = newMainModel(m.cfg, nil, nil, nil, "", nil)
	m.main.contactsSec.contacts = []contactEntry{{PubKey: "pk1", DisplayName: "Alice"}}
	m.main.contactsSec.configChanged = true

	m2, _ := m.syncMainConfig()
	if len(m2.cfg.Contacts) != 1 || m2.cfg.Contacts[0].DisplayName != "Alice" {
		t.Errorf("cfg.Contacts not synced: %+v", m2.cfg.Contacts)
	}
	if m2.main.contactsSec.configChanged {
		t.Error("configChanged flag should be cleared after sync")
	}
	if len(m2.main.contacts) != 1 {
		t.Errorf("main.contacts (chat substitution list) not updated: %+v", m2.main.contacts)
	}
}

// TestMainModel_View_HeaderRuleIsContinuous verifies the header rule does not break across the
// │ divider: the rule row is a single unbroken line, while body rows keep the divider.
func TestMainModel_View_HeaderRuleIsContinuous(t *testing.T) {
	m := makeMainModelWithChat()
	m.focus = focusLeft

	v := m.view(80, 12)
	lines := strings.Split(v, "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got %d:\n%s", len(lines), v)
	}
	// Line index 1 is the header rule row (header is line 0).
	if strings.Contains(lines[1], "│") {
		t.Errorf("header rule row should not contain the │ divider, got:\n%q", lines[1])
	}
	if !strings.Contains(lines[1], "─") {
		t.Errorf("header rule row should be a horizontal rule, got:\n%q", lines[1])
	}
	// A body row must still contain the divider.
	bodyHasDivider := false
	for _, l := range lines[2:] {
		if strings.Contains(l, "│") {
			bodyHasDivider = true
			break
		}
	}
	if !bodyHasDivider {
		t.Errorf("body rows should keep the │ divider, got:\n%s", v)
	}
}
