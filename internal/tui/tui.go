package tui

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

type model struct {
	screen     screen
	cfg        appConfig
	width      int
	height     int
	id         *identity // nil if no identity configured
	servers    serverModel
	contacts   contactsModel
	identities identitiesModel
	main       mainModel
	ident      identityModel

	currentClient *generated.ClientWithResponses
	currentServer serverConfig
}

func initialModel(cfg appConfig) model {
	m := model{cfg: cfg}

	// Restore identity from config
	if len(cfg.Identities) > 0 {
		idx := cfg.ActiveIndex
		if idx < 0 || idx >= len(cfg.Identities) {
			idx = 0
		}
		e := cfg.Identities[idx]
		id, err := identityFromHex(e.PrivateKey)
		if err == nil {
			m.id = &id
		}
	} else if cfg.Identity != nil {
		// Legacy fallback (before migration)
		id, err := identityFromHex(cfg.Identity.PrivateKey)
		if err == nil {
			m.id = &id
		}
	}

	if len(cfg.Identities) == 0 && cfg.Identity == nil {
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
			// If we just came from identity setup, save the new identity
			if prev == screenIdentity && m.ident.result.privKey != nil {
				id := m.ident.result
				m.id = &id
				m.cfg.Identities = append(m.cfg.Identities, identityEntry{
					PrivateKey: id.PrivKeyHex,
					PublicKey:  id.PubKeyHex,
				})
				m.cfg.ActiveIndex = len(m.cfg.Identities) - 1
				_ = saveConfig(m.cfg)
			}
			m.servers = newServerModel(m.cfg)

		case screenIdentities:
			m.identities = newIdentitiesModel(m.cfg)

		case screenContacts:
			m.contacts = newContactsModel(m.cfg)

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
				if npub := m.id.NpubKey; len(npub) >= 6 {
					username = npub[len(npub)-6:]
				} else {
					username = m.id.PubKeyHex[:12] + "…"
				}
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
		if m.servers.configChanged {
			m.cfg.Servers = m.servers.servers
			m.servers.configChanged = false
			_ = saveConfig(m.cfg)
		}
		return m, cmd

	case screenIdentities:
		var cmd tea.Cmd
		m.identities, cmd = m.identities.update(msg)
		if m.identities.configChanged {
			m.cfg.Identities = m.identities.entries
			m.cfg.ActiveIndex = m.identities.activeIndex
			m.identities.configChanged = false
			// Update active identity in memory
			if m.cfg.ActiveIndex < len(m.cfg.Identities) {
				e := m.cfg.Identities[m.cfg.ActiveIndex]
				if id, err := identityFromHex(e.PrivateKey); err == nil {
					m.id = &id
				}
			}
			_ = saveConfig(m.cfg)
		}
		return m, cmd

	case screenContacts:
		var cmd tea.Cmd
		m.contacts, cmd = m.contacts.update(msg)
		if m.contacts.configChanged {
			m.cfg.Contacts = m.contacts.contacts
			m.contacts.configChanged = false
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
	case screenIdentities:
		content = m.identities.view(m.width, m.height)
	case screenContacts:
		content = m.contacts.view(m.width, m.height)
	case screenRooms:
		content = m.main.view(m.width, m.height)
	default:
		content = "\n\n  µchat\n"
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func Run() error {
	if err := checkConfigPermissions(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: "+err.Error())
		return err
	}
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: could not load config:", err)
		return err
	}

	p := tea.NewProgram(initialModel(cfg))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
