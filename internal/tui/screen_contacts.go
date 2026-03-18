package tui

import (
	"strings"
	"unicode/utf8"

	table "charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

type contactsState int

const (
	contactsStateList    contactsState = iota
	contactsStateAddNpub               // typing npub
	contactsStateAddName               // typing display name
	contactsStateRename                // editing display name of selected contact
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
	case tea.PasteMsg:
		switch m.state {
		case contactsStateAddNpub:
			m.inputNpub += msg.Content
		case contactsStateAddName:
			m.inputName += msg.Content
		case contactsStateRename:
			m.inputName += msg.Content
		}

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
			case "r":
				if len(m.contacts) > 0 {
					m.state = contactsStateRename
					m.inputName = m.contacts[m.cursor].DisplayName
					m.err = ""
				}
			case "d":
				if len(m.contacts) > 0 {
					m.contacts = append(m.contacts[:m.cursor], m.contacts[m.cursor+1:]...)
					if m.cursor > 0 && m.cursor >= len(m.contacts) {
						m.cursor = len(m.contacts) - 1
					}
					m.configChanged = true
				}
			case "esc", "tab":
				return m, func() tea.Msg { return navigateMsg{to: screenRooms} }
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
				if _, size := utf8.DecodeLastRuneInString(m.inputNpub); size > 0 {
					m.inputNpub = m.inputNpub[:len(m.inputNpub)-size]
				}
			case "esc":
				m.state = contactsStateList
				m.inputNpub = ""
				m.err = ""
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if utf8.RuneCountInString(s) == 1 {
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
				if _, size := utf8.DecodeLastRuneInString(m.inputName); size > 0 {
					m.inputName = m.inputName[:len(m.inputName)-size]
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
				if utf8.RuneCountInString(s) == 1 {
					m.inputName += s
				}
			}

		case contactsStateRename:
			switch msg.String() {
			case "enter":
				m.contacts[m.cursor].DisplayName = strings.TrimSpace(m.inputName)
				m.configChanged = true
				m.state = contactsStateList
				m.inputName = ""
			case "backspace":
				if _, size := utf8.DecodeLastRuneInString(m.inputName); size > 0 {
					m.inputName = m.inputName[:len(m.inputName)-size]
				}
			case "esc":
				m.state = contactsStateList
				m.inputName = ""
			case "ctrl+c":
				return m, tea.Quit
			default:
				if s := msg.String(); utf8.RuneCountInString(s) == 1 {
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
			nameW, keyW := 4, 10
			rows := make([]table.Row, len(m.contacts))
			for i, c := range m.contacts {
				r, g, bv := pubkeyColor(c.PubKey)
				name := ansiColor(c.DisplayName, r, g, bv)
				displayKey := c.PubKey
				if npub, err := pubKeyHexToNpub(c.PubKey); err == nil {
					displayKey = npub
				}
				key := formatKeyFull(displayKey)
				rows[i] = table.Row{name, key}
				if w := visibleWidth(name); w > nameW {
					nameW = w
				}
				if w := visibleWidth(key); w > keyW {
					keyW = w
				}
			}
			cols := []table.Column{
				{Title: "Name", Width: max(nameW, visibleWidth("Name"))},
				{Title: "Public Key", Width: max(keyW, visibleWidth("Public Key"))},
			}
			b.WriteString(renderTable(cols, rows, m.cursor, pad))
		}
		b.WriteString("\n")
		b.WriteString(helpBar("↑↓", "navigate", "a", "add", "r", "rename", "d", "delete", "esc", "back", "q", "quit") + "\n")

	case contactsStateAddNpub:
		b.WriteString(pad + "Enter npub (public key):\n\n")
		b.WriteString(pad + "> " + m.inputNpub + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")

	case contactsStateAddName:
		b.WriteString(pad + "Enter display name:\n\n")
		b.WriteString(pad + "> " + m.inputName + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")

	case contactsStateRename:
		b.WriteString(pad + "Rename contact:\n\n")
		b.WriteString(pad + "> " + m.inputName + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")
	}

	if m.err != "" {
		b.WriteString("\n" + pad + "Error: " + m.err + "\n")
	}

	return b.String()
}
