package tui

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
	"github.com/mattn/go-runewidth"
)

const (
	maxServerNameWidth = 12
	maxRoomNameWidth   = 20
)

// roomNavTargets are the management screens pinned at the bottom of the rooms panel.
// They occupy cursor indices immediately after the room list.
var roomNavTargets = []struct {
	label string
	to    screen
}{
	{"Servers", screenServers},
	{"Identities", screenIdentities},
	{"Contacts", screenContacts},
}

type roomState int

const (
	roomStateList    roomState = iota
	roomStateSearch            // typing a search query
	roomStateCreate            // typing a new room name
	roomStateLoading           // waiting for rooms list
)

// serverRoom pairs a room with the server it belongs to.
type serverRoom struct {
	server serverConfig
	room   generated.Room
}

// serverRoomsLoadedMsg carries the fetched room list for one server (or an error).
type serverRoomsLoadedMsg struct {
	serverURL string
	rooms     []generated.Room
	err       error
}

// roomCreatedMsg is sent after creating a room.
type roomCreatedMsg struct {
	room *generated.Room
	err  error
}

// roomSelectedMsg is sent when the user opens a room.
type roomSelectedMsg struct {
	server   serverConfig
	room     string
	password string
	preview  bool // true = auto-preview, don't shift focus to right panel
}

type roomModel struct {
	state          roomState
	clients        map[string]*generated.ClientWithResponses // keyed by srv.URL
	servers        []serverConfig
	serverRooms    []serverRoom
	loading        map[string]bool
	selectedServer serverConfig // for password prompt
	cursor         int
	inputText      string
	err            string
	selectedRoom   string
	roomPassword   string
	promptPasswd   bool
	passwdInput    string
}

func newRoomModel(clients map[string]*generated.ClientWithResponses, servers []serverConfig) roomModel {
	loading := make(map[string]bool, len(servers))
	for _, srv := range servers {
		loading[srv.URL] = true
	}
	state := roomStateLoading
	if len(servers) == 0 {
		state = roomStateList
	}
	return roomModel{
		clients: clients,
		servers: servers,
		loading: loading,
		state:   state,
	}
}

func (m roomModel) init() tea.Cmd {
	if len(m.servers) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(m.servers))
	for _, srv := range m.servers {
		cmds = append(cmds, m.fetchServerRooms(srv))
	}
	return tea.Batch(cmds...)
}

func (m roomModel) fetchServerRooms(srv serverConfig) tea.Cmd {
	client := m.clients[srv.URL]
	serverURL := srv.URL
	return func() tea.Msg {
		if client == nil {
			return serverRoomsLoadedMsg{serverURL: serverURL, err: fmt.Errorf("no client for %s", serverURL)}
		}
		resp, err := client.GETapiroomsWithResponse(context.Background(), nil)
		if err != nil {
			return serverRoomsLoadedMsg{serverURL: serverURL, err: err}
		}
		if resp.JSON200 == nil {
			return serverRoomsLoadedMsg{serverURL: serverURL, err: fmt.Errorf("server error: %d", resp.StatusCode())}
		}
		return serverRoomsLoadedMsg{serverURL: serverURL, rooms: *resp.JSON200}
	}
}

func (m roomModel) fetchSearch(srv serverConfig, query string) tea.Cmd {
	client := m.clients[srv.URL]
	serverURL := srv.URL
	return func() tea.Msg {
		if client == nil {
			return serverRoomsLoadedMsg{serverURL: serverURL, err: fmt.Errorf("no client for %s", serverURL)}
		}
		resp, err := client.GETapiroomssearchWithResponse(context.Background(), &generated.GETapiroomssearchParams{Q: &query})
		if err != nil {
			return serverRoomsLoadedMsg{serverURL: serverURL, err: err}
		}
		if resp.JSON200 == nil {
			return serverRoomsLoadedMsg{serverURL: serverURL, err: fmt.Errorf("search failed: %d", resp.StatusCode())}
		}
		return serverRoomsLoadedMsg{serverURL: serverURL, rooms: *resp.JSON200}
	}
}

func (m roomModel) createRoom(name string) tea.Cmd {
	if len(m.servers) == 0 {
		return func() tea.Msg { return roomCreatedMsg{err: fmt.Errorf("no server configured")} }
	}
	srv := m.servers[0]
	client := m.clients[srv.URL]
	return func() tea.Msg {
		if client == nil {
			return roomCreatedMsg{err: fmt.Errorf("no client for %s", srv.URL)}
		}
		resp, err := client.POSTapiroomsWithResponse(context.Background(), nil, generated.CreateRoomRequest{Name: name})
		if err != nil {
			return roomCreatedMsg{err: err}
		}
		if resp.JSON200 == nil {
			return roomCreatedMsg{err: fmt.Errorf("create failed: %d", resp.StatusCode())}
		}
		return roomCreatedMsg{room: resp.JSON200}
	}
}

func (m roomModel) previewCmd() tea.Cmd {
	if navIdx := m.cursor - len(m.serverRooms); navIdx >= 0 {
		// Cursor is on a pinned nav item: preview that section in the right pane.
		if navIdx < len(roomNavTargets) {
			to := roomNavTargets[navIdx].to
			return func() tea.Msg { return sectionSelectedMsg{to: to, focus: false} }
		}
		return nil
	}
	sr := m.serverRooms[m.cursor]
	if sr.room.HasPassword != nil && *sr.room.HasPassword {
		return nil // password rooms require explicit Enter
	}
	name := ""
	if sr.room.Name != nil {
		name = *sr.room.Name
	}
	srv := sr.server
	return func() tea.Msg { return roomSelectedMsg{server: srv, room: name, password: "", preview: true} }
}

func (m roomModel) findServer(serverURL string) serverConfig {
	for _, srv := range m.servers {
		if srv.URL == serverURL {
			return srv
		}
	}
	return serverConfig{URL: serverURL}
}

// roomLine formats a single room entry for the list panel.
func roomLine(sr serverRoom, cursor string) string {
	srvName := runewidth.Truncate(serverDisplayName(sr.server), maxServerNameWidth, "")
	prefix := dim(srvName + "~")
	name := "(unnamed)"
	if sr.room.Name != nil {
		name = *sr.room.Name
	}
	name = runewidth.Truncate(name, maxRoomNameWidth, "")
	lock := ""
	if sr.room.HasPassword != nil && *sr.room.HasPassword {
		lock = " 🔒"
	}
	return " " + cursor + prefix + name + lock + "\n"
}

// serverDisplayName returns the quickname if set, otherwise the hostname from the URL.
func serverDisplayName(srv serverConfig) string {
	if srv.Quickname != "" {
		return srv.Quickname
	}
	u, err := url.Parse(srv.URL)
	if err != nil || u.Hostname() == "" {
		return srv.URL
	}
	return u.Hostname()
}

func (m roomModel) update(msg tea.Msg) (roomModel, tea.Cmd) {
	// Handle password prompt overlay
	if m.promptPasswd {
		if km, ok := msg.(tea.KeyMsg); ok {
			switch km.String() {
			case "enter":
				m.roomPassword = m.passwdInput
				m.promptPasswd = false
				m.passwdInput = ""
				room := m.selectedRoom
				password := m.roomPassword
				srv := m.selectedServer
				return m, func() tea.Msg { return roomSelectedMsg{server: srv, room: room, password: password} }
			case "backspace":
				if _, size := utf8.DecodeLastRuneInString(m.passwdInput); size > 0 {
					m.passwdInput = m.passwdInput[:len(m.passwdInput)-size]
				}
			case "esc":
				m.promptPasswd = false
				m.passwdInput = ""
			default:
				s := km.String()
				if utf8.RuneCountInString(s) == 1 {
					m.passwdInput += s
				}
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case serverRoomsLoadedMsg:
		delete(m.loading, msg.serverURL)
		if msg.err != nil {
			if m.err == "" {
				m.err = msg.err.Error()
			}
		} else {
			// Rebuild serverRooms: keep entries from other servers, replace for this server
			filtered := make([]serverRoom, 0, len(m.serverRooms))
			for _, sr := range m.serverRooms {
				if sr.server.URL != msg.serverURL {
					filtered = append(filtered, sr)
				}
			}
			srv := m.findServer(msg.serverURL)
			for _, rm := range msg.rooms {
				filtered = append(filtered, serverRoom{server: srv, room: rm})
			}
			m.serverRooms = filtered
			m.err = ""
		}
		if len(m.loading) == 0 {
			m.state = roomStateList
		}
		return m, m.previewCmd()

	case roomCreatedMsg:
		m.state = roomStateList
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.err = ""
		if len(m.servers) > 0 {
			return m, m.fetchServerRooms(m.servers[0])
		}
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case roomStateList:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, m.previewCmd()
			case "down", "j":
				if m.cursor < len(m.serverRooms)+len(roomNavTargets)-1 {
					m.cursor++
				}
				return m, m.previewCmd()
			case "g":
				m.cursor = 0
				return m, m.previewCmd()
			case "G":
				if len(m.serverRooms) > 0 {
					m.cursor = len(m.serverRooms) - 1
				}
				return m, m.previewCmd()
			case "enter":
				if m.cursor < len(m.serverRooms) {
					sr := m.serverRooms[m.cursor]
					m.selectedServer = sr.server
					if sr.room.Name != nil {
						m.selectedRoom = *sr.room.Name
					}
					m.roomPassword = ""
					if sr.room.HasPassword != nil && *sr.room.HasPassword {
						m.promptPasswd = true
						m.passwdInput = ""
						return m, nil
					}
					room := m.selectedRoom
					srv := m.selectedServer
					return m, func() tea.Msg { return roomSelectedMsg{server: srv, room: room, password: ""} }
				}
				// Cursor is on a pinned nav item: open the section in the right pane and focus it.
				if navIdx := m.cursor - len(m.serverRooms); navIdx >= 0 && navIdx < len(roomNavTargets) {
					to := roomNavTargets[navIdx].to
					return m, func() tea.Msg { return sectionSelectedMsg{to: to, focus: true} }
				}
			case "/":
				m.state = roomStateSearch
				m.inputText = ""
				m.err = ""
			case "c":
				m.state = roomStateCreate
				m.inputText = ""
				m.err = ""
			case "r":
				if len(m.servers) > 0 {
					for _, srv := range m.servers {
						m.loading[srv.URL] = true
					}
					m.state = roomStateLoading
					cmds := make([]tea.Cmd, 0, len(m.servers))
					for _, srv := range m.servers {
						cmds = append(cmds, m.fetchServerRooms(srv))
					}
					return m, tea.Batch(cmds...)
				}
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case roomStateSearch:
			switch msg.String() {
			case "enter":
				m.state = roomStateLoading
				if len(m.servers) > 0 {
					return m, m.fetchSearch(m.servers[0], m.inputText)
				}
			case "backspace":
				if _, size := utf8.DecodeLastRuneInString(m.inputText); size > 0 {
					m.inputText = m.inputText[:len(m.inputText)-size]
				}
			case "esc":
				m.state = roomStateList
				m.inputText = ""
				// Restore full list for servers[0]
				if len(m.servers) > 0 {
					return m, m.fetchServerRooms(m.servers[0])
				}
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if utf8.RuneCountInString(s) == 1 {
					m.inputText += s
				}
			}

		case roomStateCreate:
			switch msg.String() {
			case "enter":
				name := strings.TrimSpace(m.inputText)
				if name == "" {
					m.err = "Room name cannot be empty"
					return m, nil
				}
				m.state = roomStateLoading
				return m, m.createRoom(name)
			case "backspace":
				if _, size := utf8.DecodeLastRuneInString(m.inputText); size > 0 {
					m.inputText = m.inputText[:len(m.inputText)-size]
				}
			case "esc":
				m.state = roomStateList
				m.inputText = ""
				m.err = ""
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if utf8.RuneCountInString(s) == 1 {
					m.inputText += s
				}
			}

		case roomStateLoading:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m roomModel) viewPanel(width, height int, focused bool) string {
	var b strings.Builder

	focusMark := " "
	if focused {
		focusMark = "*"
	}
	b.WriteString(focusMark + " µchat\n")
	b.WriteString(strings.Repeat("─", width) + "\n")

	if m.promptPasswd {
		b.WriteString(" Password required:\n")
		masked := strings.Repeat("*", len(m.passwdInput))
		b.WriteString(" > " + masked + "█\n")
		b.WriteString("\n")
		b.WriteString(helpBar("enter", "ok", "esc", "cancel") + "\n")
		return b.String()
	}

	// Build the body lines for the current state; nav is pinned in list/loading states.
	var body []string
	if m.err != "" {
		body = append(body, " Error: "+m.err)
	}
	showNav := false
	switch m.state {
	case roomStateLoading:
		showNav = true
		if len(m.serverRooms) == 0 {
			body = append(body, " Loading…")
		} else {
			// Partial results: show rooms from servers that already responded
			for i, sr := range m.serverRooms {
				body = append(body, roomCursorLine(sr, i == m.cursor))
			}
			body = append(body, dim(" (loading…)"))
		}
	case roomStateList:
		showNav = true
		if len(m.serverRooms) == 0 {
			body = append(body, " (no rooms)")
			if len(m.servers) == 0 {
				body = append(body, " [tab] to add a server")
			} else {
				body = append(body, " [c] to create one")
			}
		} else {
			for i, sr := range m.serverRooms {
				body = append(body, roomCursorLine(sr, i == m.cursor))
			}
		}
	case roomStateSearch:
		body = append(body, " Search: "+m.inputText+"█")
	case roomStateCreate:
		body = append(body, " New room name:", " > "+m.inputText+"█")
	}

	if !showNav {
		for _, line := range body {
			b.WriteString(line + "\n")
		}
		return b.String()
	}

	// Pinned nav block: a dim rule followed by the management screens.
	nav := make([]string, 0, len(roomNavTargets)+1)
	nav = append(nav, dim(strings.Repeat("─", width)))
	for i, item := range roomNavTargets {
		cursor := "  "
		if m.cursor == len(m.serverRooms)+i {
			cursor = "> "
		}
		nav = append(nav, " "+cursor+"⚙ "+item.label)
	}

	// Pin nav to the bottom: header took 2 lines; fill the gap between body and nav.
	avail := max(height-2-len(nav), 0)
	if len(body) > avail {
		body = body[:avail] // keep the nav visible on very small terminals
	}
	for _, line := range body {
		b.WriteString(line + "\n")
	}
	for i := len(body); i < avail; i++ {
		b.WriteString("\n")
	}
	for _, line := range nav {
		b.WriteString(line + "\n")
	}

	return b.String()
}

// roomCursorLine renders a room entry with the selection caret, trimming the trailing newline.
func roomCursorLine(sr serverRoom, selected bool) string {
	cursor := "  "
	if selected {
		cursor = "> "
	}
	return strings.TrimRight(roomLine(sr, cursor), "\n")
}
