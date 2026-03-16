package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

// messagesLoadedMsg carries fetched messages or an error.
type messagesLoadedMsg struct {
	messages []generated.Message
	err      error
}

// messageSentMsg is sent after posting a message.
type messageSentMsg struct {
	err error
}

type chatModel struct {
	client   *generated.ClientWithResponses
	room     string
	password string
	id       *identity
	username string

	messages  []generated.Message
	inputText string
	err       string
	loading   bool
	scroll    int // offset from the bottom (0 = latest)
}

func newChatModel(client *generated.ClientWithResponses, room, password string, id *identity, username string) chatModel {
	return chatModel{
		client:   client,
		room:     room,
		password: password,
		id:       id,
		username: username,
		loading:  true,
	}
}

func (m chatModel) init() tea.Cmd {
	return m.fetchMessages()
}

func (m chatModel) fetchMessages() tea.Cmd {
	client := m.client
	room := m.room
	password := m.password
	return func() tea.Msg {
		params := &generated.GETapiroomsRoommessagesParams{}
		if password != "" {
			params.Password = &password
		}
		resp, err := client.GETapiroomsRoommessagesWithResponse(context.Background(), room, params)
		if err != nil {
			return messagesLoadedMsg{err: err}
		}
		if resp.JSON200 == nil {
			return messagesLoadedMsg{err: fmt.Errorf("server error: %d", resp.StatusCode())}
		}
		return messagesLoadedMsg{messages: *resp.JSON200}
	}
}

func (m chatModel) sendMessage(content string) tea.Cmd {
	client := m.client
	room := m.room
	password := m.password
	id := m.id
	username := m.username
	return func() tea.Msg {
		req := generated.SendMessageRequest{
			Content: content,
			User:    username,
		}
		if password != "" {
			req.RoomPassword = &password
		}
		if id != nil {
			ts := time.Now().Unix()
			sig, err := id.SignMessage(content, room, ts)
			if err == nil {
				req.Pubkey = &id.PubKeyHex
				req.Signature = &sig
				req.Timestamp = &ts
			}
		}
		resp, err := client.POSTapiroomsRoommessagesWithResponse(context.Background(), room, nil, req)
		if err != nil {
			return messageSentMsg{err: err}
		}
		if resp.JSON200 == nil {
			return messageSentMsg{err: fmt.Errorf("send failed: %d", resp.StatusCode())}
		}
		return messageSentMsg{}
	}
}

func (m chatModel) update(msg tea.Msg) (chatModel, tea.Cmd) {
	switch msg := msg.(type) {
	case messagesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.messages = msg.messages
		m.scroll = 0
		m.err = ""
		return m, nil

	case messageSentMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.err = ""
		// Refresh messages after sending
		return m, m.fetchMessages()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			content := strings.TrimSpace(m.inputText)
			if content == "" || m.username == "" {
				if m.username == "" {
					m.err = "Set a username in Identity screen first"
				}
				return m, nil
			}
			m.inputText = ""
			m.err = ""
			return m, m.sendMessage(content)
		case "backspace":
			if len(m.inputText) > 0 {
				m.inputText = m.inputText[:len(m.inputText)-1]
			}
		case "up":
			m.scroll++
		case "down":
			if m.scroll > 0 {
				m.scroll--
			}
		case "r":
			m.loading = true
			return m, m.fetchMessages()
		case "ctrl+c":
			return m, tea.Quit
		default:
			s := msg.String()
			if len(s) == 1 {
				m.inputText += s
			}
		}
	}
	return m, nil
}

func (m chatModel) viewPanel(width, height int, focused bool) string {
	var b strings.Builder

	focusMark := " "
	if focused {
		focusMark = "*"
	}
	b.WriteString(focusMark + " #" + m.room + "\n")
	b.WriteString(strings.Repeat("─", width) + "\n")

	// Reserve: header(1) + sep(1) + bottom_sep(1) + input(1) + footer(1) = 5
	contentHeight := height - 5
	if contentHeight < 1 {
		contentHeight = 1
	}

	if m.loading {
		b.WriteString(" Loading messages…\n")
	} else if len(m.messages) == 0 {
		b.WriteString(" (no messages yet)\n")
	} else {
		start := len(m.messages) - contentHeight - m.scroll
		if start < 0 {
			start = 0
		}
		end := start + contentHeight
		if end > len(m.messages) {
			end = len(m.messages)
		}
		// Pad empty lines above messages so they appear at the bottom
		shown := end - start
		for i := 0; i < contentHeight-shown; i++ {
			b.WriteString("\n")
		}
		for _, msg := range m.messages[start:end] {
			user := "?"
			if msg.User != nil && *msg.User != "" {
				user = *msg.User
			} else if msg.Pubkey != nil {
				pk := *msg.Pubkey
				if len(pk) > 8 {
					pk = pk[:8]
				}
				user = pk + "…"
			}
			content := ""
			if msg.Content != nil {
				content = *msg.Content
			}
			b.WriteString(fmt.Sprintf(" <%s> %s\n", user, content))
		}
	}

	b.WriteString(strings.Repeat("─", width) + "\n")
	if m.username != "" {
		b.WriteString(" " + m.username + "> " + m.inputText + "█\n")
	} else {
		b.WriteString(" (no username)> " + m.inputText + "█\n")
	}
	if m.err != "" {
		b.WriteString(" Err: " + m.err + "\n")
	} else {
		b.WriteString(" [Enter] Send  [↑↓] Scroll  [Tab] Panels\n")
	}

	return b.String()
}
