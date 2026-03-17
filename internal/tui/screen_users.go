package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type usersState int

const (
	usersStateList    usersState = iota
	usersStateAddNpub            // typing npub
	usersStateAddName            // typing display name
)

type usersModel struct {
	state         usersState
	users         []userEntry
	cursor        int
	inputNpub     string
	inputName     string
	err           string
	configChanged bool
}

func newUsersModel(cfg appConfig) usersModel {
	return usersModel{users: cfg.Users}
}

func (m usersModel) update(msg tea.Msg) (usersModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case usersStateList:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.users)-1 {
					m.cursor++
				}
			case "a":
				m.state = usersStateAddNpub
				m.inputNpub = ""
				m.inputName = ""
				m.err = ""
			case "d":
				if len(m.users) > 0 {
					m.users = append(m.users[:m.cursor], m.users[m.cursor+1:]...)
					if m.cursor > 0 && m.cursor >= len(m.users) {
						m.cursor = len(m.users) - 1
					}
					m.configChanged = true
				}
			case "esc", "tab":
				return m, func() tea.Msg { return navigateMsg{to: screenServers} }
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case usersStateAddNpub:
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.inputNpub) == "" {
					m.err = "npub cannot be empty"
					return m, nil
				}
				m.state = usersStateAddName
				m.err = ""
			case "backspace":
				if len(m.inputNpub) > 0 {
					m.inputNpub = m.inputNpub[:len(m.inputNpub)-1]
				}
			case "esc":
				m.state = usersStateList
				m.inputNpub = ""
				m.err = ""
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if len(s) == 1 {
					m.inputNpub += s
				}
			}

		case usersStateAddName:
			switch msg.String() {
			case "enter":
				m.users = append(m.users, userEntry{
					PubKey:      strings.TrimSpace(m.inputNpub),
					DisplayName: strings.TrimSpace(m.inputName),
				})
				m.configChanged = true
				m.state = usersStateList
				m.cursor = len(m.users) - 1
				m.inputNpub = ""
				m.inputName = ""
				m.err = ""
			case "backspace":
				if len(m.inputName) > 0 {
					m.inputName = m.inputName[:len(m.inputName)-1]
				}
			case "esc":
				m.state = usersStateList
				m.inputNpub = ""
				m.inputName = ""
				m.err = ""
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if len(s) == 1 {
					m.inputName += s
				}
			}
		}
	}
	return m, nil
}

func (m usersModel) view(width, height int) string {
	var b strings.Builder
	pad := "  "

	b.WriteString("\n\n")
	b.WriteString(pad + "µchat — Contacts\n\n")

	switch m.state {
	case usersStateList:
		if len(m.users) == 0 {
			b.WriteString(pad + "(no contacts — press [a] to add one)\n")
		} else {
			for i, u := range m.users {
				cursor := "  "
				if i == m.cursor {
					cursor = "▶ "
				}
				name := u.DisplayName
				if name == "" {
					name = u.PubKey
				}
				b.WriteString(fmt.Sprintf("%s%s%s  %s\n", pad, cursor, name, u.PubKey))
			}
		}
		b.WriteString("\n")
		b.WriteString(helpBar("↑↓", "navigate", "a", "add", "d", "delete", "esc", "back", "q", "quit") + "\n")

	case usersStateAddNpub:
		b.WriteString(pad + "Enter npub (public key):\n\n")
		b.WriteString(pad + "> " + m.inputNpub + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")

	case usersStateAddName:
		b.WriteString(pad + "Enter display name:\n\n")
		b.WriteString(pad + "> " + m.inputName + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")
	}

	if m.err != "" {
		b.WriteString("\n" + pad + "Error: " + m.err + "\n")
	}

	return b.String()
}
