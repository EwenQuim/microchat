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
}

func buildClientsMap(servers []serverConfig) map[string]*generated.ClientWithResponses {
	clients := make(map[string]*generated.ClientWithResponses, len(servers))
	for _, srv := range servers {
		u := srv.URL
		if !strings.Contains(u, "http") {
			u = "https://" + u
		}
		if !strings.HasSuffix(u, "/") {
			u += "/"
		}
		client, err := generated.NewClientWithResponses(u)
		if err != nil {
			continue // skip invalid URLs
		}
		clients[srv.URL] = client
	}
	return clients
}

func deriveUsername(id *identity, cfg appConfig) string {
	if id == nil {
		return ""
	}
	username := ""
	if cfg.ActiveIndex < len(cfg.Identities) {
		username = cfg.Identities[cfg.ActiveIndex].Name
	}
	if username == "" {
		if npub := id.NpubKey; len(npub) >= 6 {
			username = npub[len(npub)-6:]
		} else {
			username = id.PubKeyHex[:12] + "…"
		}
	}
	return username
}

// syncMainConfig persists and propagates config edits made in the in-pane management
// sections of mainModel, mirroring the per-screen save logic used for the full-screen
// settings screens. It clears the changed flags it consumes.
func (m model) syncMainConfig() (model, tea.Cmd) {
	var cmd tea.Cmd

	if m.main.serversSec.configChanged {
		m.main.serversSec.configChanged = false
		m.cfg.Servers = m.main.serversSec.servers
		_ = saveConfig(m.cfg)
		// Refresh the rooms list against the updated server set, preserving the cursor.
		clients := buildClientsMap(m.cfg.Servers)
		m.main.clients = clients
		m.main.servers = m.cfg.Servers
		m.main.rooms.clients = clients
		m.main.rooms.servers = m.cfg.Servers
		m.main.rooms.loading = make(map[string]bool, len(m.cfg.Servers))
		for _, srv := range m.cfg.Servers {
			m.main.rooms.loading[srv.URL] = true
		}
		cmd = tea.Batch(cmd, m.main.rooms.init())
	}

	if m.main.identitiesSec.configChanged {
		m.main.identitiesSec.configChanged = false
		m.cfg.Identities = m.main.identitiesSec.entries
		m.cfg.ActiveIndex = m.main.identitiesSec.activeIndex
		if m.cfg.ActiveIndex >= 0 && m.cfg.ActiveIndex < len(m.cfg.Identities) {
			if id, err := identityFromHex(m.cfg.Identities[m.cfg.ActiveIndex].PrivateKey); err == nil {
				m.id = &id
				m.main.id = &id
				m.main.username = deriveUsername(&id, m.cfg)
				m.main.chat.id = &id
				m.main.chat.username = m.main.username
			}
		}
		_ = saveConfig(m.cfg)
	}

	if m.main.contactsSec.configChanged {
		m.main.contactsSec.configChanged = false
		m.cfg.Contacts = m.main.contactsSec.contacts
		m.main.contacts = m.cfg.Contacts
		m.main.chat.contacts = m.cfg.Contacts
		_ = saveConfig(m.cfg)
	}

	return m, cmd
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

	// Always init servers model
	m.servers = newServerModel(cfg)

	if len(cfg.Identities) == 0 && cfg.Identity == nil {
		m.screen = screenIdentity
		m.ident = newIdentityModel()
	} else {
		servers := m.servers.servers
		clients := buildClientsMap(servers)
		username := deriveUsername(m.id, cfg)
		m.screen = screenRooms
		m.main = newMainModel(cfg, clients, servers, m.id, username, cfg.Contacts)
	}

	return m
}

func (m model) Init() tea.Cmd {
	if m.screen == screenRooms {
		return m.main.init()
	}
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
			servers := m.servers.servers
			clients := buildClientsMap(servers)
			username := deriveUsername(m.id, m.cfg)
			m.main = newMainModel(m.cfg, clients, servers, m.id, username, m.cfg.Contacts)
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
		if ac, ok := msg.(addContactFromChatMsg); ok {
			entry := contactEntry{PubKey: ac.pubKeyHex, DisplayName: ac.displayName}
			found := false
			for i, c := range m.cfg.Contacts {
				if c.PubKey == entry.PubKey {
					m.cfg.Contacts[i] = entry
					found = true
					break
				}
			}
			if !found {
				m.cfg.Contacts = append(m.cfg.Contacts, entry)
			}
			_ = saveConfig(m.cfg)
			m.contacts = newContactsModel(m.cfg)
			m.main.contacts = m.cfg.Contacts
			m.main.chat.contacts = m.cfg.Contacts
			return m, nil
		}
		var cmd tea.Cmd
		m.main, cmd = m.main.update(msg)
		m, cmd2 := m.syncMainConfig()
		return m, tea.Batch(cmd, cmd2)
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
