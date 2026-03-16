package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

type roomState int

const (
	roomStateList    roomState = iota
	roomStateSearch            // typing a search query
	roomStateCreate            // typing a new room name
	roomStateLoading           // waiting for rooms list
)

// roomsLoadedMsg carries the fetched room list (or an error).
type roomsLoadedMsg struct {
	rooms []generated.Room
	err   error
}

// roomCreatedMsg is sent after creating a room.
type roomCreatedMsg struct {
	room *generated.Room
	err  error
}

// roomSelectedMsg is sent when the user opens a room.
type roomSelectedMsg struct {
	room     string
	password string
	preview  bool // true = auto-preview, don't shift focus to right panel
}

type roomModel struct {
	state        roomState
	client       *generated.ClientWithResponses
	rooms        []generated.Room
	cursor       int
	inputText    string
	err          string
	selectedRoom string
	roomPassword string
	promptPasswd bool
	passwdInput  string
}

func newRoomModel(client *generated.ClientWithResponses) roomModel {
	return roomModel{client: client, state: roomStateLoading}
}

func (m roomModel) init() tea.Cmd {
	return m.fetchRooms("")
}

func (m roomModel) fetchRooms(query string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
		if query != "" {
			resp, err := client.GETapiroomssearchWithResponse(context.Background(), &generated.GETapiroomssearchParams{Q: &query})
			if err != nil {
				return roomsLoadedMsg{err: err}
			}
			if resp.JSON200 == nil {
				return roomsLoadedMsg{err: fmt.Errorf("search failed: %d", resp.StatusCode())}
			}
			return roomsLoadedMsg{rooms: *resp.JSON200}
		}
		resp, err := client.GETapiroomsWithResponse(context.Background(), nil)
		if err != nil {
			return roomsLoadedMsg{err: err}
		}
		if resp.JSON200 == nil {
			return roomsLoadedMsg{err: fmt.Errorf("server error: %d", resp.StatusCode())}
		}
		return roomsLoadedMsg{rooms: *resp.JSON200}
	}
}

func (m roomModel) createRoom(name string) tea.Cmd {
	client := m.client
	return func() tea.Msg {
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
	if m.cursor >= len(m.rooms) {
		return nil
	}
	rm := m.rooms[m.cursor]
	if rm.HasPassword != nil && *rm.HasPassword {
		return nil // password rooms require explicit Enter
	}
	name := ""
	if rm.Name != nil {
		name = *rm.Name
	}
	return func() tea.Msg { return roomSelectedMsg{room: name, password: "", preview: true} }
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
				return m, func() tea.Msg { return roomSelectedMsg{room: room, password: password} }
			case "backspace":
				if len(m.passwdInput) > 0 {
					m.passwdInput = m.passwdInput[:len(m.passwdInput)-1]
				}
			case "esc":
				m.promptPasswd = false
				m.passwdInput = ""
			default:
				s := km.String()
				if len(s) == 1 {
					m.passwdInput += s
				}
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case roomsLoadedMsg:
		m.state = roomStateList
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.rooms = msg.rooms
		m.cursor = 0
		m.err = ""
		return m, nil

	case roomCreatedMsg:
		m.state = roomStateList
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.err = ""
		// Refresh room list
		return m, m.fetchRooms("")

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
				if m.cursor < len(m.rooms)-1 {
					m.cursor++
				}
				return m, m.previewCmd()
			case "enter":
				if m.cursor < len(m.rooms) {
					rm := m.rooms[m.cursor]
					if rm.Name != nil {
						m.selectedRoom = *rm.Name
					}
					m.roomPassword = ""
					if rm.HasPassword != nil && *rm.HasPassword {
						m.promptPasswd = true
						m.passwdInput = ""
						return m, nil
					}
					room := m.selectedRoom
					return m, func() tea.Msg { return roomSelectedMsg{room: room, password: ""} }
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
				m.state = roomStateLoading
				return m, m.fetchRooms("")
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case roomStateSearch:
			switch msg.String() {
			case "enter":
				m.state = roomStateLoading
				return m, m.fetchRooms(m.inputText)
			case "backspace":
				if len(m.inputText) > 0 {
					m.inputText = m.inputText[:len(m.inputText)-1]
				}
			case "esc":
				m.state = roomStateList
				m.inputText = ""
				// Restore full list
				return m, m.fetchRooms("")
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if len(s) == 1 {
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
				if len(m.inputText) > 0 {
					m.inputText = m.inputText[:len(m.inputText)-1]
				}
			case "esc":
				m.state = roomStateList
				m.inputText = ""
				m.err = ""
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if len(s) == 1 {
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
	b.WriteString(focusMark + " Rooms\n")
	b.WriteString(strings.Repeat("─", width) + "\n")

	if m.promptPasswd {
		b.WriteString(" Password required:\n")
		masked := strings.Repeat("*", len(m.passwdInput))
		b.WriteString(" > " + masked + "█\n")
		b.WriteString("\n")
		b.WriteString(" [Enter] OK  [Esc] Cancel\n")
		return b.String()
	}

	if m.err != "" {
		b.WriteString(" Error: " + m.err + "\n")
	}

	switch m.state {
	case roomStateLoading:
		b.WriteString(" Loading…\n")
	case roomStateList:
		if len(m.rooms) == 0 {
			b.WriteString(" (no rooms)\n")
			b.WriteString(" [c] to create one\n")
		} else {
			for i, rm := range m.rooms {
				cursor := "  "
				if i == m.cursor {
					cursor = "▶ "
				}
				name := "(unnamed)"
				if rm.Name != nil {
					name = *rm.Name
				}
				lock := ""
				if rm.HasPassword != nil && *rm.HasPassword {
					lock = " 🔒"
				}
				b.WriteString(" " + cursor + name + lock + "\n")
			}
		}
	case roomStateSearch:
		b.WriteString(" Search: " + m.inputText + "█\n")
	case roomStateCreate:
		b.WriteString(" New room name:\n")
		b.WriteString(" > " + m.inputText + "█\n")
	}

	return b.String()
}
