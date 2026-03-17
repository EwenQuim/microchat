package tui

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
)

type idState int

const (
	idStateMenu            idState = iota
	idStateInput                   // pasting a private key
	idStateVanityInput             // typing the vanity suffix
	idStateVanityGenerating        // goroutines running
)

type idMode int

const (
	idModeSetup idMode = iota // first-run: emits navigateMsg{to: screenServers}
	idModeAdd                 // add to list: emits identityCreatedMsg
)

type vanityProgressMsg struct{ attempts int64 }
type vanityFoundMsg struct{ id identity }
type identityCreatedMsg struct{ id identity }

type identityModel struct {
	state         idState
	mode          idMode
	inputText     string
	err           string
	result        identity
	vanityCancel  context.CancelFunc
	vanityCounter *atomic.Int64
	vanityInput   string // suffix being searched
}

func newIdentityModel() identityModel {
	return identityModel{state: idStateMenu, mode: idModeSetup}
}

func newIdentityModelAdd() identityModel {
	return identityModel{state: idStateMenu, mode: idModeAdd}
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
				if m.mode == idModeAdd {
					id := m.result
					return m, func() tea.Msg { return identityCreatedMsg{id: id} }
				}
				return m, func() tea.Msg { return navigateMsg{to: screenServers} }
			case "p":
				m.state = idStateInput
				m.inputText = ""
				m.err = ""
			case "v":
				m.state = idStateVanityInput
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
				if m.mode == idModeAdd {
					id := m.result
					return m, func() tea.Msg { return identityCreatedMsg{id: id} }
				}
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

		case idStateVanityInput:
			switch msg.String() {
			case "enter":
				if m.inputText == "" {
					m.err = "Enter 1–5 bech32 characters"
					return m, nil
				}
				suffix := m.inputText
				m.vanityInput = suffix
				m.vanityCounter = &atomic.Int64{}
				ctx, cancel := context.WithCancel(context.Background())
				m.vanityCancel = cancel
				m.state = idStateVanityGenerating
				m.err = ""
				counter := m.vanityCounter
				return m, tea.Batch(
					startVanityCmd(ctx, suffix, counter),
					vanityTickCmd(counter),
				)
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
				if len(s) == 1 && isValidBech32Char(s[0]) && len(m.inputText) < 5 {
					m.inputText += s
				} else if len(s) == 1 && isValidBech32Char(s[0]) {
					m.err = "Max 5 chars in TUI. Use the CLI with --unsafe-cpu-usage for longer suffixes."
				}
			}

		case idStateVanityGenerating:
			switch msg.String() {
			case "esc", "ctrl+c":
				if m.vanityCancel != nil {
					m.vanityCancel()
					m.vanityCancel = nil
				}
				m.state = idStateMenu
				m.inputText = ""
				m.err = ""
			}
		}

	case vanityFoundMsg:
		if m.vanityCancel != nil {
			m.vanityCancel()
			m.vanityCancel = nil
		}
		m.result = msg.id
		m.err = ""
		if m.mode == idModeAdd {
			id := m.result
			return m, func() tea.Msg { return identityCreatedMsg{id: id} }
		}
		return m, func() tea.Msg { return navigateMsg{to: screenServers} }

	case vanityProgressMsg:
		if m.state == idStateVanityGenerating {
			counter := m.vanityCounter
			return m, vanityTickCmd(counter)
		}
	}
	return m, nil
}

func startVanityCmd(ctx context.Context, suffix string, counter *atomic.Int64) tea.Cmd {
	return func() tea.Msg {
		id, err := generateVanityIdentity(ctx, suffix, counter)
		if err != nil {
			// context cancelled — return a progress msg as a safe no-op
			return vanityProgressMsg{attempts: counter.Load()}
		}
		return vanityFoundMsg{id: id}
	}
}

func vanityTickCmd(counter *atomic.Int64) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(100 * time.Millisecond)
		return vanityProgressMsg{attempts: counter.Load()}
	}
}

func (m identityModel) view(width, height int) string {
	var b strings.Builder
	pad := strings.Repeat(" ", 2)

	b.WriteString("\n\n")
	b.WriteString(pad + "µchat — Identity Setup\n\n")

	switch m.state {
	case idStateMenu:
		b.WriteString(helpBar("g", "generate new keypair", "p", "paste private key", "v", "vanity keypair", "q", "quit") + "\n")
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

	case idStateVanityInput:
		b.WriteString(pad + "Enter vanity suffix (1–5 bech32 chars, e.g. cafe):\n\n")
		b.WriteString(pad + "> " + m.inputText + "█\n\n")
		b.WriteString(helpBar("enter", "start", "esc", "cancel") + "\n")
		if m.err != "" {
			b.WriteString(fmt.Sprintf("\n%s  Error: %s\n", pad, m.err))
		}

	case idStateVanityGenerating:
		attempts := int64(0)
		if m.vanityCounter != nil {
			attempts = m.vanityCounter.Load()
		}
		b.WriteString(fmt.Sprintf("%sSearching for npub ending in %q…\n\n", pad, m.vanityInput))
		b.WriteString(fmt.Sprintf("%s%s attempts\n\n", pad, FormatCount(attempts)))
		b.WriteString(helpBar("esc", "cancel") + "\n")
	}

	return b.String()
}
