package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
	"github.com/mattn/go-runewidth"
)

type mainFocus int

const (
	focusLeft  mainFocus = iota
	focusRight
)

type mainModel struct {
	focus    mainFocus
	rooms    roomModel
	chat     chatModel
	hasChat  bool
	client   *generated.ClientWithResponses
	id       *identity
	username string
}

func newMainModel(client *generated.ClientWithResponses, id *identity, username string) mainModel {
	return mainModel{
		client:   client,
		id:       id,
		username: username,
		rooms:    newRoomModel(client),
		focus:    focusLeft,
	}
}

func (m mainModel) init() tea.Cmd {
	return m.rooms.init()
}

func (m mainModel) update(msg tea.Msg) (mainModel, tea.Cmd) {
	switch msg := msg.(type) {
	case roomSelectedMsg:
		m.chat = newChatModel(m.client, msg.room, msg.password, m.id, m.username)
		m.hasChat = true
		if !msg.preview {
			m.focus = focusRight
		}
		return m, m.chat.init()

	case roomsLoadedMsg, roomCreatedMsg:
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
			if m.hasChat {
				if m.focus == focusLeft {
					m.focus = focusRight
				} else {
					m.focus = focusLeft
				}
			}
			return m, nil
		case "esc":
			if m.focus == focusRight {
				m.focus = focusLeft
				return m, nil
			}
			return m, func() tea.Msg { return navigateMsg{to: screenServers} }
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
	leftWidth := 28
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
	for i := 0; i < height; i++ {
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
		b.WriteString(leftPadded + "│" + rightStr + "\n")
	}
	return b.String()
}

func dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}

func padRight(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw >= width {
		return runewidth.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-sw)
}
