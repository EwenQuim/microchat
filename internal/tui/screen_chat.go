package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
	colorful "github.com/lucasb-eyer/go-colorful"
)

func pubkeyColor(pubkey string) (r, g, b uint8) {
	var hash int32
	for _, ch := range pubkey {
		hash = int32(ch) + ((hash << 5) - hash)
	}
	if hash < 0 {
		hash = -hash
	}
	hue := hash % 360
	c := colorful.Hsl(float64(hue), 0.70, 0.55)
	return uint8(c.R * 255), uint8(c.G * 255), uint8(c.B * 255)
}

func ansiColor(s string, r, g, b uint8) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, s)
}

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

	messages   []generated.Message
	inputText  string
	err        string
	loading    bool
	scroll     int  // offset from the bottom (0 = latest)
	typing     bool // vim-style insert mode
	colorCache map[string][3]uint8
}

func newChatModel(client *generated.ClientWithResponses, room, password string, id *identity, username string) chatModel {
	return chatModel{
		client:     client,
		room:       room,
		password:   password,
		id:         id,
		username:   username,
		loading:    true,
		colorCache: make(map[string][3]uint8),
	}
}

func (m chatModel) cachedColor(key string) (r, g, b uint8) {
	if c, ok := m.colorCache[key]; ok {
		return c[0], c[1], c[2]
	}
	r, g, b = pubkeyColor(key)
	m.colorCache[key] = [3]uint8{r, g, b}
	return r, g, b
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
		if m.typing {
			switch msg.String() {
			case "esc", "tab":
				m.typing = false
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
			default:
				if t := msg.Key().Text; t != "" {
					m.inputText += t
				}
			}
		} else {
			switch msg.String() {
			case "i":
				m.typing = true
			case "enter":
				content := strings.TrimSpace(m.inputText)
				if content != "" && m.username != "" {
					m.inputText = ""
					m.err = ""
					return m, m.sendMessage(content)
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
	sep := strings.Repeat("─", width)
	b.WriteString(focusMark + " #" + m.room + "\n")
	b.WriteString(sep + "\n")

	// Reserve: header(1) + sep(1) + bottom_sep(1) + input(1) + footer(1) = 5
	contentHeight := max(height-5, 1)

	if m.loading {
		for i := 0; i < contentHeight-1; i++ {
			b.WriteString("\n")
		}
		b.WriteString(" Loading messages…\n")
	} else if len(m.messages) == 0 {
		for i := 0; i < contentHeight-1; i++ {
			b.WriteString("\n")
		}
		b.WriteString(" (no messages yet)\n")
	} else {
		start := max(len(m.messages)-contentHeight-m.scroll, 0)
		end := min(start+contentHeight, len(m.messages))
		// Pad empty lines above messages so they appear at the bottom
		shown := end - start
		for i := 0; i < contentHeight-shown; i++ {
			b.WriteString("\n")
		}
		for _, msg := range m.messages[start:end] {
			var fullPk, truncPk string
			if msg.Pubkey != nil && *msg.Pubkey != "" {
				fullPk = *msg.Pubkey
				npub, err := pubKeyHexToNpub(fullPk)
				if err == nil && len(npub) >= 8 {
					truncPk = npub[len(npub)-8:]
				}
			}
			user := "?"
			if msg.User != nil && *msg.User != "" {
				user = *msg.User
			} else if truncPk != "" {
				user = truncPk + "…"
			}
			content := ""
			if msg.Content != nil {
				content = *msg.Content
			}
			keyLabel := ""
			if truncPk != "" {
				keyLabel = " " + dim(truncPk)
			}
			colorKey := user
			if fullPk != "" {
				colorKey = fullPk
			}
			r, g, bv := m.cachedColor(colorKey)
			coloredUser := ansiColor(user, r, g, bv)
			fmt.Fprintf(&b, " %s%s%s %s\n", coloredUser, keyLabel, dim(":"), content)
		}
	}

	b.WriteString(sep + "\n")
	cursor := ""
	if m.typing {
		cursor = "█"
	}
	if m.username != "" && m.id != nil {
		r, g, bv := m.cachedColor(m.id.PubKeyHex)
		coloredName := ansiColor(m.username, r, g, bv)

		truncPk := m.id.PubKeyHex
		if pk := m.id.PubKeyHex; len(pk) >= 8 {
			npub, err := pubKeyHexToNpub(pk)
			if err == nil && len(npub) >= 8 {
				truncPk = npub[len(npub)-8:]
			}
		}
		b.WriteString(" " + coloredName + " " + dim(truncPk) + "> " + m.inputText + cursor + "\n")
	} else if m.username != "" {
		b.WriteString(" " + m.username + "> " + m.inputText + cursor + "\n")
	} else {
		b.WriteString(" (no username)> " + m.inputText + cursor + "\n")
	}
	if m.err != "" {
		b.WriteString(" Err: " + m.err + "\n")
	} else if m.typing {
		b.WriteString(helpBar("esc", "exit", "enter", "send", "⌫", "delete") + "\n")
	} else {
		b.WriteString(helpBar("i", "insert", "r", "refresh", "↑↓", "scroll", "tab", "panels") + "\n")
	}

	return b.String()
}
