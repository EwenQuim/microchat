package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type idState int

const (
	idStateMenu  idState = iota
	idStateInput         // pasting a private key
)

type identityModel struct {
	state     idState
	inputText string
	err       string
	result    identity
}

func newIdentityModel() identityModel {
	return identityModel{state: idStateMenu}
}

func (m identityModel) update(msg tea.Msg) (identityModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case idStateMenu:
			switch msg.String() {
			case "g":
				id, err := generateIdentity()
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.result = id
				m.err = ""
				return m, func() tea.Msg { return navigateMsg{to: screenServers} }
			case "p":
				m.state = idStateInput
				m.inputText = ""
				m.err = ""
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case idStateInput:
			switch msg.String() {
			case "enter":
				id, err := identityFromHex(strings.TrimSpace(m.inputText))
				if err != nil {
					m.err = "Invalid private key (expected 64 hex chars)"
					m.state = idStateMenu
					m.inputText = ""
					return m, nil
				}
				m.result = id
				m.err = ""
				return m, func() tea.Msg { return navigateMsg{to: screenServers} }
			case "backspace":
				if len(m.inputText) > 0 {
					m.inputText = m.inputText[:len(m.inputText)-1]
				}
			case "esc", "ctrl+c":
				m.state = idStateMenu
				m.inputText = ""
				m.err = ""
			default:
				s := msg.String()
				if len(s) == 1 {
					m.inputText += s
				}
			}
		}
	}
	return m, nil
}

func (m identityModel) view(width, height int) string {
	var b strings.Builder
	pad := strings.Repeat(" ", 2)

	b.WriteString("\n\n")
	b.WriteString(pad + "µchat — Identity Setup\n\n")

	switch m.state {
	case idStateMenu:
		b.WriteString(helpBar("g", "generate new keypair", "p", "paste private key", "q", "quit") + "\n")
		if m.err != "" {
			b.WriteString(fmt.Sprintf("\n%s  Error: %s\n", pad, m.err))
		}

	case idStateInput:
		b.WriteString(pad + "Paste your private key (hex):\n\n")
		b.WriteString(pad + "> " + m.inputText + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")
		if m.err != "" {
			b.WriteString(fmt.Sprintf("\n%s  Error: %s\n", pad, m.err))
		}
	}

	return b.String()
}
