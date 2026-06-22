package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

func makeRoomModel(servers ...serverConfig) roomModel {
	clients := make(map[string]*generated.ClientWithResponses)
	return newRoomModel(clients, servers)
}

func makeRoom(name string) generated.Room {
	n := name
	return generated.Room{Name: &n}
}

// TestRoomModel_Init_EmitsOneCmdPerServer verifies init returns nil for zero servers,
// and a non-nil batch cmd when servers are present.
func TestRoomModel_Init_EmitsOneCmdPerServer(t *testing.T) {
	m := makeRoomModel()
	cmd := m.init()
	if cmd != nil {
		t.Error("init with no servers should return nil cmd")
	}

	m2 := makeRoomModel(
		serverConfig{URL: "http://a.example"},
		serverConfig{URL: "http://b.example"},
	)
	cmd2 := m2.init()
	if cmd2 == nil {
		t.Error("init with servers should return a non-nil cmd")
	}
}

// TestServerRoomsLoadedMsg_AggregatesAcrossServers verifies rooms accumulate across multiple servers.
func TestServerRoomsLoadedMsg_AggregatesAcrossServers(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	srvB := serverConfig{URL: "http://b.example", Quickname: "Beta"}
	m := makeRoomModel(srvA, srvB)

	roomA := makeRoom("alpha-room")
	m2, _ := m.update(serverRoomsLoadedMsg{serverURL: srvA.URL, rooms: []generated.Room{roomA}})
	if len(m2.serverRooms) != 1 {
		t.Fatalf("serverRooms count = %d, want 1 after first server loaded", len(m2.serverRooms))
	}
	if m2.serverRooms[0].server.URL != srvA.URL {
		t.Errorf("serverRooms[0].server.URL = %q, want %q", m2.serverRooms[0].server.URL, srvA.URL)
	}

	roomB := makeRoom("beta-room")
	m3, _ := m2.update(serverRoomsLoadedMsg{serverURL: srvB.URL, rooms: []generated.Room{roomB}})
	if len(m3.serverRooms) != 2 {
		t.Fatalf("serverRooms count = %d, want 2 after both servers loaded", len(m3.serverRooms))
	}
}

// TestServerRoomsLoadedMsg_TransitionsToListWhenAllLoaded verifies state transitions after all servers respond.
func TestServerRoomsLoadedMsg_TransitionsToListWhenAllLoaded(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example"}
	srvB := serverConfig{URL: "http://b.example"}
	m := makeRoomModel(srvA, srvB)

	if m.state != roomStateLoading {
		t.Errorf("initial state = %v, want roomStateLoading", m.state)
	}

	m2, _ := m.update(serverRoomsLoadedMsg{serverURL: srvA.URL, rooms: nil})
	if m2.state != roomStateLoading {
		t.Errorf("state after first response = %v, want roomStateLoading (still loading srvB)", m2.state)
	}

	m3, _ := m2.update(serverRoomsLoadedMsg{serverURL: srvB.URL, rooms: nil})
	if m3.state != roomStateList {
		t.Errorf("state after all responses = %v, want roomStateList", m3.state)
	}
}

// TestServerRoomsLoadedMsg_PartialResultsVisibleWhileLoading verifies rooms are shown while other servers still load.
func TestServerRoomsLoadedMsg_PartialResultsVisibleWhileLoading(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	srvB := serverConfig{URL: "http://b.example", Quickname: "Beta"}
	m := makeRoomModel(srvA, srvB)

	roomA := makeRoom("alpha-room")
	m2, _ := m.update(serverRoomsLoadedMsg{serverURL: srvA.URL, rooms: []generated.Room{roomA}})

	if len(m2.serverRooms) != 1 {
		t.Fatalf("serverRooms count = %d, want 1 (partial load)", len(m2.serverRooms))
	}

	v := m2.viewPanel(60, 24, true)
	if !strings.Contains(v, "alpha-room") {
		t.Errorf("view should show alpha-room during partial load, got:\n%s", v)
	}
	if !strings.Contains(v, "loading") {
		t.Errorf("view should show loading indicator while srvB is loading, got:\n%s", v)
	}
}

// TestRoomModel_ViewPanel_ShowsServerPrefix verifies rooms are displayed with server~ prefix.
func TestRoomModel_ViewPanel_ShowsServerPrefix(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	m := makeRoomModel(srvA)
	m.state = roomStateList
	m.serverRooms = []serverRoom{{server: srvA, room: makeRoom("general")}}

	v := m.viewPanel(60, 24, true)
	if !strings.Contains(v, "~") {
		t.Errorf("view should contain ~ separator, got:\n%s", v)
	}
	if !strings.Contains(v, "general") {
		t.Errorf("view should contain room name, got:\n%s", v)
	}
}

// TestServerDisplayName_UsesQuickname verifies quickname is preferred over hostname.
func TestServerDisplayName_UsesQuickname(t *testing.T) {
	srv := serverConfig{URL: "http://example.com", Quickname: "ExampleServer"}
	got := serverDisplayName(srv)
	if got != "ExampleServer" {
		t.Errorf("serverDisplayName = %q, want ExampleServer", got)
	}
}

// TestServerDisplayName_FallsBackToHostname verifies hostname is used when quickname is absent.
func TestServerDisplayName_FallsBackToHostname(t *testing.T) {
	srv := serverConfig{URL: "http://example.com"}
	got := serverDisplayName(srv)
	if got != "example.com" {
		t.Errorf("serverDisplayName = %q, want example.com", got)
	}
}

// TestRoomModel_PreviewCmd_PopulatesServerField verifies server is included in the preview message.
func TestRoomModel_PreviewCmd_PopulatesServerField(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	m := makeRoomModel(srvA)
	m.state = roomStateList
	m.serverRooms = []serverRoom{{server: srvA, room: makeRoom("general")}}
	m.cursor = 0

	cmd := m.previewCmd()
	if cmd == nil {
		t.Fatal("previewCmd should return non-nil cmd for a non-password room")
	}
	msg := runCmd(cmd)
	rsm, ok := msg.(roomSelectedMsg)
	if !ok {
		t.Fatalf("expected roomSelectedMsg, got %T", msg)
	}
	if rsm.server.URL != srvA.URL {
		t.Errorf("roomSelectedMsg.server.URL = %q, want %q", rsm.server.URL, srvA.URL)
	}
	if rsm.room != "general" {
		t.Errorf("roomSelectedMsg.room = %q, want general", rsm.room)
	}
	if !rsm.preview {
		t.Error("roomSelectedMsg.preview should be true for previewCmd")
	}
}

// TestRoomModel_ZeroServers_StartsInListState verifies that with no servers the model starts in list state.
func TestRoomModel_ZeroServers_StartsInListState(t *testing.T) {
	m := makeRoomModel()
	if m.state != roomStateList {
		t.Errorf("state = %v, want roomStateList when no servers configured", m.state)
	}
}

// TestRoomModel_ViewPanel_ShowsNavItems verifies the sidebar nav (Servers/Identities/Contacts) is rendered.
func TestRoomModel_ViewPanel_ShowsNavItems(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	m := makeRoomModel(srvA)
	m.state = roomStateList
	m.serverRooms = []serverRoom{{server: srvA, room: makeRoom("general")}}

	v := m.viewPanel(40, 24, true)
	for _, want := range []string{"Servers", "Identities", "Contacts", "⚙"} {
		if !strings.Contains(v, want) {
			t.Errorf("view should contain nav item %q, got:\n%s", want, v)
		}
	}
}

// TestRoomModel_Down_CrossesIntoNav verifies the cursor moves from the last room into the nav section.
func TestRoomModel_Down_CrossesIntoNav(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	m := makeRoomModel(srvA)
	m.state = roomStateList
	m.serverRooms = []serverRoom{{server: srvA, room: makeRoom("general")}}
	m.cursor = 0

	m2, _ := m.update(pressKey(tea.KeyDown))
	if m2.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 (first nav item) after down from last room", m2.cursor)
	}

	// Cursor must not move past the last nav item (1 room + 3 nav => max index 3).
	for range 5 {
		m2, _ = m2.update(pressKey(tea.KeyDown))
	}
	if m2.cursor != 3 {
		t.Errorf("cursor = %d, want 3 (clamped at last nav item)", m2.cursor)
	}
}

// TestRoomModel_Enter_OnNav_SelectsSection verifies Enter on a nav item emits a
// focused sectionSelectedMsg for the matching section.
func TestRoomModel_Enter_OnNav_SelectsSection(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	cases := []struct {
		cursor int
		want   screen
	}{
		{1, screenServers},
		{2, screenIdentities},
		{3, screenContacts},
	}
	for _, tc := range cases {
		m := makeRoomModel(srvA)
		m.state = roomStateList
		m.serverRooms = []serverRoom{{server: srvA, room: makeRoom("general")}}
		m.cursor = tc.cursor

		_, cmd := m.update(pressKey(tea.KeyEnter))
		msg := runCmd(cmd)
		sec, ok := msg.(sectionSelectedMsg)
		if !ok {
			t.Fatalf("cursor %d: expected sectionSelectedMsg, got %T", tc.cursor, msg)
		}
		if sec.to != tc.want {
			t.Errorf("cursor %d: sectionSelectedMsg.to = %v, want %v", tc.cursor, sec.to, tc.want)
		}
		if !sec.focus {
			t.Errorf("cursor %d: Enter on a nav item should focus the section", tc.cursor)
		}
	}
}

// TestRoomModel_PreviewCmd_OnNav_SelectsSection verifies arrowing onto a nav item
// previews that section (unfocused) in the right pane.
func TestRoomModel_PreviewCmd_OnNav_SelectsSection(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	m := makeRoomModel(srvA)
	m.state = roomStateList
	m.serverRooms = []serverRoom{{server: srvA, room: makeRoom("general")}}
	m.cursor = 1 // first nav item (Servers)

	msg := runCmd(m.previewCmd())
	sec, ok := msg.(sectionSelectedMsg)
	if !ok {
		t.Fatalf("expected sectionSelectedMsg from previewCmd on nav item, got %T", msg)
	}
	if sec.to != screenServers {
		t.Errorf("sectionSelectedMsg.to = %v, want screenServers", sec.to)
	}
	if sec.focus {
		t.Error("arrow preview should not focus the section")
	}
}

// TestRoomModel_Enter_OnRoom_StillOpensRoom verifies room selection is unaffected by the nav section.
func TestRoomModel_Enter_OnRoom_StillOpensRoom(t *testing.T) {
	srvA := serverConfig{URL: "http://a.example", Quickname: "Alpha"}
	m := makeRoomModel(srvA)
	m.state = roomStateList
	m.serverRooms = []serverRoom{{server: srvA, room: makeRoom("general")}}
	m.cursor = 0

	_, cmd := m.update(pressKey(tea.KeyEnter))
	msg := runCmd(cmd)
	if _, ok := msg.(roomSelectedMsg); !ok {
		t.Fatalf("expected roomSelectedMsg when entering a room, got %T", msg)
	}
}
