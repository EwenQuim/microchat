package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

func makeServerModel(servers ...serverConfig) serverModel {
	cfg := appConfig{Servers: servers}
	return newServerModel(cfg)
}

func TestServerModel_SelectedServer_Empty(t *testing.T) {
	m := makeServerModel()
	got := m.selectedServer()
	if got != (serverConfig{}) {
		t.Errorf("expected zero value for empty server list, got %+v", got)
	}
}

func TestServerModel_SelectedServer_ReturnsCorrect(t *testing.T) {
	a := serverConfig{URL: "http://a.example"}
	b := serverConfig{URL: "http://b.example"}
	m := makeServerModel(a, b)
	m.cursor = 1

	got := m.selectedServer()
	if got.URL != b.URL {
		t.Errorf("selectedServer = %q, want %q", got.URL, b.URL)
	}
}

func TestServerModel_CursorUp(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)
	m.cursor = 1

	m2, _ := m.update(pressKey(tea.KeyUp))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0", m2.cursor)
	}
}

func TestServerModel_CursorUp_k(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)
	m.cursor = 1

	m2, _ := m.update(pressChar("k"))
	if m2.cursor != 0 {
		t.Errorf("cursor after k = %d, want 0", m2.cursor)
	}
}

func TestServerModel_CursorUp_Bounded(t *testing.T) {
	m := makeServerModel(serverConfig{URL: "http://a.example"})
	m.cursor = 0

	m2, _ := m.update(pressKey(tea.KeyUp))
	if m2.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", m2.cursor)
	}
}

func TestServerModel_CursorDown(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)
	m.cursor = 0

	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 1 {
		t.Errorf("cursor = %d, want 1", m2.cursor)
	}
}

func TestServerModel_CursorDown_j(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)

	m2, _ := m.update(pressChar("j"))
	if m2.cursor != 1 {
		t.Errorf("cursor after j = %d, want 1", m2.cursor)
	}
}

func TestServerModel_CursorDown_Bounded(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)
	m.cursor = 1

	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 1 {
		t.Errorf("cursor should not exceed len-1, got %d", m2.cursor)
	}
}

func TestServerModel_Enter_WithServers(t *testing.T) {
	m := makeServerModel(serverConfig{URL: "http://a.example"})

	_, cmd := m.update(pressKey(tea.KeyEnter))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenRooms {
		t.Errorf("expected screenRooms, got %v", nav.to)
	}
}

func TestServerModel_Enter_NoServers(t *testing.T) {
	m := makeServerModel()

	_, cmd := m.update(pressKey(tea.KeyEnter))
	if cmd != nil {
		t.Error("expected nil cmd when no servers")
	}
}

func TestServerModel_PressA_EntersAddState(t *testing.T) {
	m := makeServerModel()

	m2, _ := m.update(pressChar("a"))
	if m2.state != serverStateAddURL {
		t.Errorf("expected serverStateAddURL, got %v", m2.state)
	}
}

func TestServerModel_AddState_TypingBuildsInput(t *testing.T) {
	m := makeServerModel()
	m.state = serverStateAddURL

	m2, _ := m.update(pressChar("h"))
	if m2.inputText != "h" {
		t.Errorf("inputText = %q, want %q", m2.inputText, "h")
	}
}

func TestServerModel_AddState_Esc(t *testing.T) {
	m := makeServerModel()
	m.state = serverStateAddURL
	m.inputText = "something"

	m2, _ := m.update(pressKey(tea.KeyEscape))
	if m2.state != serverStateList {
		t.Errorf("expected serverStateList, got %v", m2.state)
	}
	if m2.inputText != "" {
		t.Errorf("inputText should be cleared, got %q", m2.inputText)
	}
}

func TestServerModel_AddState_EnterEmpty(t *testing.T) {
	m := makeServerModel()
	m.state = serverStateAddURL
	m.inputText = ""

	m2, cmd := m.update(pressKey(tea.KeyEnter))
	if m2.err == "" {
		t.Error("expected error for empty URL")
	}
	if cmd != nil {
		t.Error("expected nil cmd for empty URL")
	}
}

func TestServerModel_AddState_EnterURL_TransitionsToLoading(t *testing.T) {
	m := makeServerModel()
	m.state = serverStateAddURL
	m.inputText = "http://example.com"

	m2, cmd := m.update(pressKey(tea.KeyEnter))
	if m2.state != serverStateLoading {
		t.Errorf("expected serverStateLoading, got %v", m2.state)
	}
	if cmd == nil {
		t.Error("expected a cmd to be returned")
	}
}

func TestServerModel_Delete_RemovesServer(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)
	m.cursor = 0

	m2, _ := m.update(pressChar("d"))
	if len(m2.servers) != 1 {
		t.Errorf("servers count = %d, want 1", len(m2.servers))
	}
	if m2.servers[0].URL != "http://b.example" {
		t.Errorf("remaining server = %q, want b.example", m2.servers[0].URL)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
}

func TestServerModel_Delete_AdjustsCursor(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)
	m.cursor = 1 // last item

	m2, _ := m.update(pressChar("d"))
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0 after deleting last item", m2.cursor)
	}
}

func TestServerModel_Delete_EmptyList(t *testing.T) {
	m := makeServerModel()

	m2, _ := m.update(pressChar("d"))
	if len(m2.servers) != 0 {
		t.Errorf("servers count = %d, want 0", len(m2.servers))
	}
}

func TestServerModel_Tab_NavigatesToIdentity(t *testing.T) {
	m := makeServerModel()

	_, cmd := m.update(pressKey(tea.KeyTab))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenIdentity {
		t.Errorf("expected screenIdentity, got %v", nav.to)
	}
}

func TestServerModel_ServerInfoMsg_Success(t *testing.T) {
	name := "TestServer"
	desc := "A test server"
	m := makeServerModel()

	info := &generated.ServerInfoResponse{
		SuggestedQuickname: &name,
		Description:        &desc,
	}
	m2, _ := m.update(serverInfoMsg{url: "http://example.com", info: info})

	if len(m2.servers) != 1 {
		t.Fatalf("servers count = %d, want 1", len(m2.servers))
	}
	srv := m2.servers[0]
	if srv.URL != "http://example.com" {
		t.Errorf("URL = %q, want http://example.com", srv.URL)
	}
	if srv.Quickname != name {
		t.Errorf("Quickname = %q, want %q", srv.Quickname, name)
	}
	if srv.Description != desc {
		t.Errorf("Description = %q, want %q", srv.Description, desc)
	}
	if m2.cursor != 0 {
		t.Errorf("cursor = %d, want 0 (last item)", m2.cursor)
	}
	if !m2.configChanged {
		t.Error("configChanged should be true")
	}
	if m2.state != serverStateList {
		t.Errorf("state = %v, want serverStateList", m2.state)
	}
}

func TestServerModel_ServerInfoMsg_Error(t *testing.T) {
	m := makeServerModel()
	m.state = serverStateLoading

	m2, _ := m.update(serverInfoMsg{url: "http://bad.example", err: fmt.Errorf("connection refused")})
	if m2.state != serverStateList {
		t.Errorf("state = %v, want serverStateList", m2.state)
	}
	if m2.err == "" {
		t.Error("expected error string to be set")
	}
	if !strings.Contains(m2.err, "http://bad.example") {
		t.Errorf("error should mention URL, got %q", m2.err)
	}
	if len(m2.servers) != 0 {
		t.Errorf("servers count = %d, want 0", len(m2.servers))
	}
}

func TestServerModel_View_ShowsCursor(t *testing.T) {
	m := makeServerModel(
		serverConfig{URL: "http://a.example", Quickname: "Alpha"},
		serverConfig{URL: "http://b.example", Quickname: "Beta"},
	)
	m.cursor = 0

	v := m.view(80, 24)
	if !strings.Contains(v, "▶") {
		t.Error("view should show cursor ▶")
	}
}

func TestServerModel_View_EmptyList(t *testing.T) {
	m := makeServerModel()

	v := m.view(80, 24)
	if !strings.Contains(v, "no servers") {
		t.Errorf("view should show no-servers message, got:\n%s", v)
	}
}

func TestServerModel_PressU_NavigatesToUsers(t *testing.T) {
	m := makeServerModel()
	_, cmd := m.update(pressChar("u"))
	msg := runCmd(cmd)
	nav, ok := msg.(navigateMsg)
	if !ok {
		t.Fatalf("expected navigateMsg, got %T", msg)
	}
	if nav.to != screenUsers {
		t.Errorf("expected screenUsers, got %v", nav.to)
	}
}
