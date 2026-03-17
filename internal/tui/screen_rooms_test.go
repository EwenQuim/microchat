package tui

import (
	"strings"
	"testing"

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
