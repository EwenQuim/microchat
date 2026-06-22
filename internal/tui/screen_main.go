package tui

import (
	"fmt"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
	"github.com/mattn/go-runewidth"
)

// ansiEscapeRe matches ANSI escape sequences (e.g. \033[2m, \033[0m, \033[38;2;r;g;bm).
var ansiEscapeRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// visibleWidth returns the printable column width of s, ignoring ANSI escape codes.
func visibleWidth(s string) int {
	return runewidth.StringWidth(ansiEscapeRe.ReplaceAllString(s, ""))
}

// isRuleLine reports whether s (ignoring ANSI codes) is a non-empty run of ─ characters,
// i.e. a horizontal separator rule.
func isRuleLine(s string) bool {
	stripped := ansiEscapeRe.ReplaceAllString(s, "")
	if stripped == "" {
		return false
	}
	for _, r := range stripped {
		if r != '─' {
			return false
		}
	}
	return true
}

type mainFocus int

const (
	focusLeft mainFocus = iota
	focusRight
)

// rightContent identifies what the right pane is currently showing.
type rightContent int

const (
	rightSelectRoom rightContent = iota // nothing selected yet
	rightChat
	rightServers
	rightIdentities
	rightContacts
)

// sectionSelectedMsg switches the right pane to a management section. focus=true also
// moves focus into the right pane (as Enter does); focus=false is a preview (stay left).
type sectionSelectedMsg struct {
	to    screen
	focus bool
}

// sectionToRight maps a navigable screen id to the matching right-pane content.
func sectionToRight(to screen) rightContent {
	switch to {
	case screenServers:
		return rightServers
	case screenIdentities:
		return rightIdentities
	case screenContacts:
		return rightContacts
	default:
		return rightSelectRoom
	}
}

type mainModel struct {
	focus mainFocus
	right rightContent
	rooms roomModel
	chat  chatModel

	// In-pane management sections (rendered in the right pane).
	serversSec    serverModel
	identitiesSec identitiesModel
	contactsSec   contactsModel

	hasChat  bool
	cfg      appConfig
	clients  map[string]*generated.ClientWithResponses
	servers  []serverConfig
	id       *identity
	username string
	contacts []contactEntry
}

func newMainModel(cfg appConfig, clients map[string]*generated.ClientWithResponses, servers []serverConfig, id *identity, username string, contacts []contactEntry) mainModel {
	return mainModel{
		cfg:           cfg,
		clients:       clients,
		servers:       servers,
		id:            id,
		username:      username,
		contacts:      contacts,
		rooms:         newRoomModel(clients, servers),
		serversSec:    newServerModel(cfg),
		identitiesSec: newIdentitiesModel(cfg),
		contactsSec:   newContactsModel(cfg),
		focus:         focusLeft,
	}
}

// isConfigContent reports whether the right pane currently shows a management section.
func isConfigContent(r rightContent) bool {
	return r == rightServers || r == rightIdentities || r == rightContacts
}

// rightEditing reports whether the active config section is in a text-input / busy
// sub-state (so keys must be delegated raw rather than intercepted for navigation).
func (m mainModel) rightEditing() bool {
	switch m.right {
	case rightServers:
		return m.serversSec.state != serverStateList
	case rightIdentities:
		return m.identitiesSec.state != identitiesStateList
	case rightContacts:
		return m.contactsSec.state != contactsStateList
	}
	return false
}

// navIndexFor returns the sidebar cursor index of a section's nav item.
func (m mainModel) navIndexFor(to screen) int {
	for i, t := range roomNavTargets {
		if t.to == to {
			return len(m.rooms.serverRooms) + i
		}
	}
	return len(m.rooms.serverRooms)
}

func (m mainModel) init() tea.Cmd {
	return m.rooms.init()
}

func (m mainModel) update(msg tea.Msg) (mainModel, tea.Cmd) {
	switch msg := msg.(type) {
	case roomSelectedMsg:
		client := m.clients[msg.server.URL]
		m.chat = newChatModel(client, msg.server, msg.room, msg.password, m.id, m.username)
		m.chat.contacts = m.contacts
		m.hasChat = true
		m.right = rightChat
		if !msg.preview {
			m.focus = focusRight
		}
		return m, m.chat.init()

	case sectionSelectedMsg:
		m.right = sectionToRight(msg.to)
		// Rebuild the section from the latest config so it reflects saved state.
		switch msg.to {
		case screenServers:
			m.serversSec = newServerModel(m.cfg)
		case screenIdentities:
			m.identitiesSec = newIdentitiesModel(m.cfg)
		case screenContacts:
			m.contactsSec = newContactsModel(m.cfg)
		}
		if msg.focus {
			m.focus = focusRight
		}
		return m, nil

	case serverRoomsLoadedMsg, roomCreatedMsg:
		var cmd tea.Cmd
		m.rooms, cmd = m.rooms.update(msg)
		return m, cmd

	case messagesLoadedMsg, olderMessagesLoadedMsg, messageSentMsg:
		if m.hasChat {
			var cmd tea.Cmd
			m.chat, cmd = m.chat.update(msg)
			return m, cmd
		}
		return m, nil

	case serverInfoMsg:
		// Result of adding a server in the in-pane Servers section.
		var cmd tea.Cmd
		m.serversSec, cmd = m.serversSec.update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Forward any other async message to the active config section (e.g. identity
	// creation / vanity progress, paste events).
	if isConfigContent(m.right) {
		return m.delegateConfig(msg)
	}
	return m, nil
}

// handleKey routes a key press according to focus and right-pane content.
func (m mainModel) handleKey(msg tea.KeyMsg) (mainModel, tea.Cmd) {
	// Edit/typing bypass: when an input is active in the focused right pane, deliver
	// the key raw so Esc/Tab/Enter behave as that input expects.
	if m.focus == focusRight {
		if m.right == rightChat && m.hasChat && m.chat.typing {
			var cmd tea.Cmd
			m.chat, cmd = m.chat.update(msg)
			return m, cmd
		}
		if isConfigContent(m.right) && m.rightEditing() {
			return m.delegateConfig(msg)
		}
	}

	key := msg.String()

	// Global section shortcuts (preview, stay on the left).
	switch key {
	case "S", "I", "C":
		to := map[string]screen{"S": screenServers, "I": screenIdentities, "C": screenContacts}[key]
		m.rooms.cursor = m.navIndexFor(to)
		return m, func() tea.Msg { return sectionSelectedMsg{to: to, focus: false} }
	}

	hasRight := m.right != rightSelectRoom

	switch key {
	case "tab":
		if m.focus == focusLeft && hasRight {
			m.focus = focusRight
		} else if m.focus == focusRight {
			m.focus = focusLeft
		}
		return m, nil
	case "right":
		if m.focus == focusLeft && hasRight {
			m.focus = focusRight
		}
		return m, nil
	case "left":
		if m.focus == focusRight {
			m.focus = focusLeft
		}
		return m, nil
	case "esc":
		if m.focus == focusRight {
			m.focus = focusLeft
		}
		return m, nil
	}

	// Default delegation by focus + content.
	if m.focus == focusLeft {
		var cmd tea.Cmd
		m.rooms, cmd = m.rooms.update(msg)
		return m, cmd
	}
	if m.right == rightChat && m.hasChat {
		var cmd tea.Cmd
		m.chat, cmd = m.chat.update(msg)
		return m, cmd
	}
	if isConfigContent(m.right) {
		// List-state config: handle navigation keys here; delegate only safe list keys.
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		if m.right == rightServers && (key == "enter" || key == "u") {
			return m, nil // these would navigate away in serverModel; ignore in-pane
		}
		return m.delegateConfig(msg)
	}
	return m, nil
}

// delegateConfig forwards a message to the active right-pane config section.
func (m mainModel) delegateConfig(msg tea.Msg) (mainModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.right {
	case rightServers:
		m.serversSec, cmd = m.serversSec.update(msg)
	case rightIdentities:
		m.identitiesSec, cmd = m.identitiesSec.update(msg)
	case rightContacts:
		m.contactsSec, cmd = m.contactsSec.update(msg)
	}
	return m, cmd
}

func (m mainModel) view(width, height int) string {
	const roomLineMax = 1 + 2 + maxServerNameWidth + 1 + maxRoomNameWidth // = 36
	leftWidth := roomLineMax
	if width < leftWidth+20 {
		leftWidth = width / 3
	}
	rightWidth := width - leftWidth - 1 // -1 for the │ separator

	leftStr := m.rooms.viewPanel(leftWidth, height, m.focus == focusLeft)
	rightFocused := m.focus == focusRight
	var rightStr string
	switch m.right {
	case rightChat:
		if m.hasChat {
			rightStr = m.chat.viewPanel(rightWidth, height, rightFocused)
		} else {
			rightStr = " Select a room\n"
		}
	case rightServers:
		rightStr = m.serversSec.viewPanel(rightWidth, height, rightFocused)
	case rightIdentities:
		rightStr = m.identitiesSec.viewPanel(rightWidth, height, rightFocused)
	case rightContacts:
		rightStr = m.contactsSec.viewPanel(rightWidth, height, rightFocused)
	default:
		rightStr = " Select a room\n"
	}

	leftLines := strings.Split(leftStr, "\n")
	rightLines := strings.Split(rightStr, "\n")

	// Remove trailing empty element from Split when string ends with \n
	if len(leftLines) > 0 && leftLines[len(leftLines)-1] == "" {
		leftLines = leftLines[:len(leftLines)-1]
	}
	if len(rightLines) > 0 && rightLines[len(rightLines)-1] == "" {
		rightLines = rightLines[:len(rightLines)-1]
	}

	var b strings.Builder
	for i := range height {
		left := ""
		right := ""
		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(rightLines) {
			right = rightLines[i]
		}
		leftPadded := padRight(left, leftWidth)
		rightStr := right
		// A row where both panes are horizontal rules (the header separator) is
		// joined with ─ instead of │ so it reads as one unbroken line, and is left
		// undimmed since it is a shared separator rather than panel content.
		if isRuleLine(left) && isRuleLine(right) {
			fmt.Fprintf(&b, "%s─%s\n", leftPadded, rightStr)
			continue
		}
		hasRight := m.right != rightSelectRoom
		if m.focus == focusRight {
			leftPadded = dim(leftPadded)
		} else if hasRight {
			rightStr = dim(rightStr)
		}
		fmt.Fprintf(&b, "%s│%s\n", leftPadded, rightStr)
	}
	return b.String()
}

// renderPanel composes a right-pane section: a focus-marked title, a full-width rule,
// the body lines, blank fill, and a help bar pinned at the bottom. The shape mirrors
// the chat/rooms panels so the continuous header-rule merge keeps working.
func renderPanel(width, height int, focused bool, title string, body []string, help string) string {
	var b strings.Builder
	focusMark := " "
	if focused {
		focusMark = "*"
	}
	b.WriteString(focusMark + " " + title + "\n")
	b.WriteString(strings.Repeat("─", width) + "\n")

	// header(2) + body + fill + help(1) == height
	avail := max(height-2-1, 0)
	if len(body) > avail {
		body = body[:avail]
	}
	for _, line := range body {
		b.WriteString(line + "\n")
	}
	for i := len(body); i < avail; i++ {
		b.WriteString("\n")
	}
	b.WriteString(help + "\n")
	return b.String()
}

// panelBodyLines splits a rendered block into body lines, dropping the trailing newline.
func panelBodyLines(s string) []string {
	return strings.Split(strings.TrimRight(s, "\n"), "\n")
}

// truncCell truncates a plain (non-ANSI) cell value to w columns.
func truncCell(s string, w int) string {
	if w < 1 {
		w = 1
	}
	return runewidth.Truncate(s, w, "")
}

// fitColumnWidth caps a column's desired width so the table fits within paneWidth.
// otherColsTotal is the summed width (including padding) of every other column,
// including renderTable's prepended cursor column (3).
func fitColumnWidth(paneWidth, desired, otherColsTotal int) int {
	avail := max(paneWidth-otherColsTotal-2, 3) // this column's own padding
	if desired > avail {
		return avail
	}
	return desired
}

func dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}

func helpKey(key, desc string) string {
	return key + " " + dim(desc)
}

func helpBar(pairs ...string) string {
	items := make([]string, 0, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		items = append(items, helpKey(pairs[i], pairs[i+1]))
	}
	return " " + strings.Join(items, dim("  •  "))
}

// formatKeyFull renders a full key for list display: the leading chars are
// dimmed and the final 8 characters are left at normal brightness.
func formatKeyFull(key string) string {
	if len(key) <= 8 {
		return key
	}
	return dim(key[:len(key)-8]) + key[len(key)-8:]
}

func padRight(s string, width int) string {
	sw := visibleWidth(s)
	if sw > width {
		if ansiEscapeRe.MatchString(s) {
			return s // preserve existing ANSI behavior
		}
		return runewidth.Truncate(s, width, "")
	}
	if sw == width {
		return s
	}
	return s + strings.Repeat(" ", width-sw)
}
