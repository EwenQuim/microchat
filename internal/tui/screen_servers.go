package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

type serverState int

const (
	serverStateList    serverState = iota
	serverStateAddURL              // typing a URL
	serverStateLoading             // verifying URL
)

// serverInfoMsg is received after probing a server.
type serverInfoMsg struct {
	url  string
	info *generated.ServerInfoResponse
	err  error
}

type serverModel struct {
	state         serverState
	servers       []serverConfig
	cursor        int
	inputText     string
	err           string
	configChanged bool
}

var defaultServers = []serverConfig{
	{URL: "https://microchat.go-fuego.dev"},
}

func newServerModel(cfg appConfig) serverModel {
	servers := cfg.Servers
	if len(servers) == 0 {
		servers = defaultServers
	}
	return serverModel{servers: servers}
}

func (m serverModel) selectedServer() serverConfig {
	if len(m.servers) == 0 || m.cursor >= len(m.servers) {
		return serverConfig{}
	}
	return m.servers[m.cursor]
}

func fetchServerInfo(url string) tea.Cmd {
	return func() tea.Msg {
		// Normalize URL: add scheme and trailing slash for the generated client
		serverURL := url
		if !strings.Contains(serverURL, "http") {
			serverURL = "https://" + serverURL
		}
		if !strings.HasSuffix(serverURL, "/") {
			serverURL += "/"
		}
		client, err := generated.NewClientWithResponses(serverURL)
		if err != nil {
			return serverInfoMsg{url: serverURL, err: err}
		}
		resp, err := client.GETapiserverInfoWithResponse(context.Background(), nil)
		if err != nil {
			return serverInfoMsg{url: serverURL, err: err}
		}
		if resp.JSON200 == nil {
			return serverInfoMsg{url: serverURL, err: fmt.Errorf("server returned %d", resp.StatusCode())}
		}
		return serverInfoMsg{url: serverURL, info: resp.JSON200}
	}
}

func (m serverModel) update(msg tea.Msg) (serverModel, tea.Cmd) {
	switch msg := msg.(type) {
	case serverInfoMsg:
		m.state = serverStateList
		if msg.err != nil {
			m.err = fmt.Sprintf("Cannot reach %s: %s", msg.url, msg.err)
			return m, nil
		}
		srv := serverConfig{URL: msg.url}
		if msg.info.SuggestedQuickname != nil {
			srv.Quickname = *msg.info.SuggestedQuickname
		}
		if msg.info.Description != nil {
			srv.Description = *msg.info.Description
		}
		m.servers = append(m.servers, srv)
		m.cursor = len(m.servers) - 1
		m.configChanged = true
		m.err = ""
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case serverStateList:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.servers)-1 {
					m.cursor++
				}
			case "enter":
				if len(m.servers) > 0 {
					return m, func() tea.Msg { return navigateMsg{to: screenRooms} }
				}
			case "a":
				m.state = serverStateAddURL
				m.inputText = ""
				m.err = ""
			case "d":
				if len(m.servers) > 0 {
					m.servers = append(m.servers[:m.cursor], m.servers[m.cursor+1:]...)
					if m.cursor > 0 && m.cursor >= len(m.servers) {
						m.cursor = len(m.servers) - 1
					}
					m.configChanged = true
				}
			case "u":
				return m, func() tea.Msg { return navigateMsg{to: screenContacts} }
			case "esc":
				return m, func() tea.Msg { return navigateMsg{to: screenRooms} }
			case "tab":
				return m, func() tea.Msg { return navigateMsg{to: screenIdentities} }
			case "ctrl+c", "q":
				return m, tea.Quit
			}

		case serverStateAddURL:
			switch msg.String() {
			case "enter":
				url := strings.TrimSpace(m.inputText)
				if url == "" {
					m.err = "URL cannot be empty"
					return m, nil
				}
				m.state = serverStateLoading
				m.err = ""
				return m, fetchServerInfo(url)
			case "backspace":
				if len(m.inputText) > 0 {
					m.inputText = m.inputText[:len(m.inputText)-1]
				}
			case "esc":
				m.state = serverStateList
				m.inputText = ""
				m.err = ""
			case "ctrl+c":
				return m, tea.Quit
			default:
				s := msg.String()
				if len(s) == 1 {
					m.inputText += s
				}
			}

		case serverStateLoading:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m serverModel) view(width, height int) string {
	var b strings.Builder
	pad := "  "

	b.WriteString("\n\n")
	b.WriteString(pad + "µchat — Servers\n\n")

	switch m.state {
	case serverStateList:
		if len(m.servers) == 0 {
			b.WriteString(pad + "(no servers — press [a] to add one)\n")
		} else {
			for i, srv := range m.servers {
				cursor := "  "
				if i == m.cursor {
					cursor = "▶ "
				}
				name := srv.Quickname
				if name == "" {
					name = srv.URL
				}
				fmt.Fprintf(&b, "%s%s%s  %s\n", pad, cursor, name, srv.URL)
			}
		}
		b.WriteString("\n")
		b.WriteString(helpBar("↑↓", "navigate", "enter", "open", "a", "add", "d", "delete", "u", "contacts", "tab", "identities", "esc", "rooms", "q", "quit") + "\n")

	case serverStateAddURL:
		b.WriteString(pad + "Enter server URL:\n\n")
		b.WriteString(pad + "> " + m.inputText + "█\n\n")
		b.WriteString(helpBar("enter", "confirm", "esc", "cancel") + "\n")

	case serverStateLoading:
		b.WriteString(pad + "Connecting to " + m.inputText + "…\n")
	}

	if m.err != "" {
		b.WriteString("\n" + pad + "Error: " + m.err + "\n")
	}

	return b.String()
}
