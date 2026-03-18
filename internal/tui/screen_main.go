package tui

import (
	"fmt"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
	"github.com/mattn/go-runewidth"
)

// ansiEscapeRe matches ANSI escape sequences (e.g. \033[2m, \033[0m, \033[38;2;r;g;bm).
var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// visibleWidth returns the printable column width of s, ignoring ANSI escape codes.
func visibleWidth(s string) int {
	return runewidth.StringWidth(ansiEscapeRe.ReplaceAllString(s, ""))
}

type mainFocus int

const (
	focusLeft mainFocus = iota
	focusRight
)

type mainModel struct {
	focus    mainFocus
	rooms    roomModel
	chat     chatModel
	hasChat  bool
	clients  map[string]*generated.ClientWithResponses
	servers  []serverConfig
	id       *identity
	username string
	contacts []contactEntry
}

func newMainModel(clients map[string]*generated.ClientWithResponses, servers []serverConfig, id *identity, username string, contacts []contactEntry) mainModel {
	return mainModel{
		clients:  clients,
		servers:  servers,
		id:       id,
		username: username,
		contacts: contacts,
		rooms:    newRoomModel(clients, servers),
		focus:    focusLeft,
	}
}

func (m mainModel) init() tea.Cmd {
	return m.rooms.init()
}

func (m mainModel) update(msg tea.Msg) (mainModel, tea.Cmd) {
	switch msg := msg.(type) {
	case roomSelectedMsg:
		client := m.clients[msg.server.URL]
		m.chat = newChatModel(client, msg.server, msg.room, msg.password, m.id, m.username)
		m.chat.contacts = m.contacts
		m.hasChat = true
		if !msg.preview {
			m.focus = focusRight
		}
		return m, m.chat.init()

	case serverRoomsLoadedMsg, roomCreatedMsg:
		var cmd tea.Cmd
		m.rooms, cmd = m.rooms.update(msg)
		return m, cmd

	case messagesLoadedMsg, messageSentMsg:
		if m.hasChat {
			var cmd tea.Cmd
			m.chat, cmd = m.chat.update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		// When chat is focused and in typing mode, bypass global key handlers
		// so Esc/Tab exit typing mode instead of switching panels or navigating
		if m.focus == focusRight && m.hasChat && m.chat.typing {
			var cmd tea.Cmd
			m.chat, cmd = m.chat.update(msg)
			return m, cmd
		}
		switch msg.String() {
		case "tab":
			return m, func() tea.Msg { return navigateMsg{to: screenServers} }
		case "esc":
			if m.focus == focusRight {
				m.focus = focusLeft
				return m, nil
			}
			return m, func() tea.Msg { return navigateMsg{to: screenServers} }
		case "right":
			if m.focus == focusLeft && m.hasChat {
				m.focus = focusRight
				return m, nil
			}
		case "left":
			if m.focus == focusRight {
				m.focus = focusLeft
				return m, nil
			}
		default:
			if m.focus == focusLeft {
				var cmd tea.Cmd
				m.rooms, cmd = m.rooms.update(msg)
				return m, cmd
			}
			if m.hasChat {
				var cmd tea.Cmd
				m.chat, cmd = m.chat.update(msg)
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m mainModel) view(width, height int) string {
	const roomLineMax = 1 + 2 + maxServerNameWidth + 1 + maxRoomNameWidth // = 36
	leftWidth := roomLineMax
	if width < leftWidth+20 {
		leftWidth = width / 3
	}
	rightWidth := width - leftWidth - 1 // -1 for the │ separator

	leftStr := m.rooms.viewPanel(leftWidth, height, m.focus == focusLeft)
	var rightStr string
	if m.hasChat {
		rightStr = m.chat.viewPanel(rightWidth, height, m.focus == focusRight)
	} else {
		rightStr = " Select a room\n"
	}

	leftLines := strings.Split(leftStr, "\n")
	rightLines := strings.Split(rightStr, "\n")

	// Remove trailing empty element from Split when string ends with \n
	if len(leftLines) > 0 && leftLines[len(leftLines)-1] == "" {
		leftLines = leftLines[:len(leftLines)-1]
	}
	if len(rightLines) > 0 && rightLines[len(rightLines)-1] == "" {
		rightLines = rightLines[:len(rightLines)-1]
	}

	var b strings.Builder
	for i := range height {
		left := ""
		right := ""
		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(rightLines) {
			right = rightLines[i]
		}
		leftPadded := padRight(left, leftWidth)
		rightStr := right
		if m.focus == focusRight {
			leftPadded = dim(leftPadded)
		} else if m.hasChat {
			rightStr = dim(rightStr)
		}
		fmt.Fprintf(&b, "%s│%s\n", leftPadded, rightStr)
	}
	return b.String()
}

func dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}

func helpKey(key, desc string) string {
	return key + " " + dim(desc)
}

func helpBar(pairs ...string) string {
	items := make([]string, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		items = append(items, helpKey(pairs[i], pairs[i+1]))
	}
	return " " + strings.Join(items, dim("  •  "))
}

// formatKeyFull renders a full key for list display: the leading chars are
// dimmed and the final 8 characters are left at normal brightness.
func formatKeyFull(key string) string {
	if len(key) <= 8 {
		return key
	}
	return dim(key[:len(key)-8]) + key[len(key)-8:]
}

func padRight(s string, width int) string {
	sw := visibleWidth(s)
	if sw > width {
		if ansiEscapeRe.MatchString(s) {
			return s // preserve existing ANSI behavior
		}
		return runewidth.Truncate(s, width, "")
	}
	if sw == width {
		return s
	}
	return s + strings.Repeat(" ", width-sw)
}
