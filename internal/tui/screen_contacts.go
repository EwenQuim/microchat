package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type contactsState int

const (
	contactsStateList    contactsState = iota
	contactsStateAddNpub               // typing npub
	contactsStateAddName               // typing display name
)

type contactsModel struct {
	state         contactsState
	contacts      []contactEntry
	cursor        int
	inputNpub     string
	inputName     string
	err           string
	configChanged bool
}

func newContactsModel(cfg appConfig) contactsModel {
	return contactsModel{contacts: cfg.Contacts}
}

func (m contactsModel) update(msg tea.Msg) (contactsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case contactsStateList:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.contacts)-1 {
					m.cursor++
				}
			case "a":
				m.state = contactsStateAddNpub
				m.inputNpub = ""
				m.inputName = ""
				m.err = ""
			case "d":
				if len(m.contacts) > 0 {
					m.contacts = append(m.contacts[:m.cursor], m.contacts[m.cursor+1:]...)
					if m.cursor > 0 && m.cursor >= len(m.contacts) {
						m.cursor = len(m.contacts) - 1
					}
					m.configChanged = true
				}
			case "esc", "tab":
				return m, func() tea.Msg { return navigateMsg{to: screenServers} }
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case contactsStateAddNpub:
			switch msg.String() {
			case "enter":
				if strings.TrimSpace(m.inputNpub) == "" {
					m.err = "npub cannot be empty"
					return m, nil
				}
				m.state = contactsStateAddName
				m.err = ""
			case "backspace":
				if len(m.inputNpub) > 0 {
					m.inputNpub = m.inputNpub[:len(m.inputNpub)-1]
				}
			case "esc":
				m.state = contactsStateList
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

		case contactsStateAddName:
			switch msg.String() {
			case "enter":
				m.contacts = append(m.contacts, contactEntry{
					PubKey:      strings.TrimSpace(m.inputNpub),
					DisplayName: strings.TrimSpace(m.inputName),
				})
				m.configChanged = true
				m.state = contactsStateList
				m.cursor = len(m.contacts) - 1
				m.inputNpub = ""
				m.inputName = ""
				m.err = ""
			case "backspace":
				if len(m.inputName) > 0 {
					m.inputName = m.inputName[:len(m.inputName)-1]
				}
			case "esc":
				m.state = contactsStateList
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

func (m contactsModel) view(width, height int) string {
	var b strings.Builder
	pad := "  "

	b.WriteString("\n\n")
	b.WriteString(pad + "µchat — Contacts\n\n")

	switch m.state {
	case contactsStateList:
		if len(m.contacts) == 0 {
			b.WriteString(pad + "(no contacts — press [a] to add one)\n")
		} else {
			for i, u := range m.contacts {
				cursor := "  "
				if i == m.cursor {
					cursor = "▶ "
				}
				r, g, bv := pubkeyColor(u.PubKey)
				var nameStr string
				if u.DisplayName != "" {
					nameStr = ansiColor(u.DisplayName, r, g, bv) + "  "
				}
				b.WriteString(fmt.Sprintf("%s%s%s%s\n", pad, cursor, nameStr, formatKeyFull(u.PubKey)))
			}
		}
		b.WriteString("\n")
		b.WriteString(helpBar("↑↓", "navigate", "a", "add", "d", "delete", "esc", "back", "q", "quit") + "\n")

	case contactsStateAddNpub:
		b.WriteString(pad + "Enter npub (public key):\n\n")
		b.WriteString(pad + "> " + m.inputNpub + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")

	case contactsStateAddName:
		b.WriteString(pad + "Enter display name:\n\n")
		b.WriteString(pad + "> " + m.inputName + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")
	}

	if m.err != "" {
		b.WriteString("\n" + pad + "Error: " + m.err + "\n")
	}

	return b.String()
}
