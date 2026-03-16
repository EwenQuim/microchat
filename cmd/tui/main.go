package main

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

type model struct {
	screen  screen
	cfg     appConfig
	width   int
	height  int
	id      *identity // nil if no identity configured
	servers serverModel
	main    mainModel
	ident   identityModel

	currentClient *generated.ClientWithResponses
	currentServer serverConfig
}

func initialModel(cfg appConfig) model {
	m := model{cfg: cfg}

	// Restore identity from config
	if cfg.Identity != nil {
		id, err := identityFromHex(cfg.Identity.PrivateKey)
		if err == nil {
			m.id = &id
		}
	}

	if cfg.Identity == nil {
		m.screen = screenIdentity
		m.ident = newIdentityModel()
	} else {
		m.screen = screenServers
		m.servers = newServerModel(cfg)
	}

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle global messages first
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case navigateMsg:
		prev := m.screen
		m.screen = msg.to

		switch msg.to {
		case screenIdentity:
			m.ident = newIdentityModel()

		case screenServers:
			// If we just came from identity, save the new identity
			if prev == screenIdentity && m.ident.result.privKey != nil {
				id := m.ident.result
				m.id = &id
				m.cfg.Identity = &identityConfig{
					PrivateKey: id.PrivKeyHex,
					PublicKey:  id.PubKeyHex,
				}
				_ = saveConfig(m.cfg)
			}
			m.servers = newServerModel(m.cfg)

		case screenRooms:
			srv := m.servers.selectedServer()
			m.currentServer = srv
			m.cfg.LastServer = srv.URL
			_ = saveConfig(m.cfg)

			url := srv.URL
			if !strings.HasSuffix(url, "/") {
				url += "/"
			}
			client, err := generated.NewClientWithResponses(url)
			if err != nil {
				// Fall back to servers screen with error
				m.screen = screenServers
				m.servers.err = fmt.Sprintf("invalid server URL: %s", err)
				return m, nil
			}
			m.currentClient = client

			username := ""
			if m.id != nil {
				username = m.id.PubKeyHex[:12] + "…"
			}
			m.main = newMainModel(client, m.id, username)
			return m, m.main.init()
		}
		return m, nil
	}

	// Delegate to active screen
	switch m.screen {
	case screenIdentity:
		var cmd tea.Cmd
		m.ident, cmd = m.ident.update(msg)
		return m, cmd

	case screenServers:
		var cmd tea.Cmd
		m.servers, cmd = m.servers.update(msg)
		// Persist config changes immediately
		if m.servers.configChanged {
			m.cfg.Servers = m.servers.servers
			m.servers.configChanged = false
			_ = saveConfig(m.cfg)
		}
		return m, cmd

	case screenRooms:
		var cmd tea.Cmd
		m.main, cmd = m.main.update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() tea.View {
	var content string
	switch m.screen {
	case screenIdentity:
		content = m.ident.view(m.width, m.height)
	case screenServers:
		content = m.servers.view(m.width, m.height)
	case screenRooms:
		content = m.main.view(m.width, m.height)
	default:
		content = "\n\n  µchat\n"
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: could not load config:", err)
	}

	p := tea.NewProgram(initialModel(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
