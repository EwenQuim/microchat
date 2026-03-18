package tui

import (
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
)

// FuzzVisibleWidth ensures visibleWidth never panics and always returns >= 0.
func FuzzVisibleWidth(f *testing.F) {
	f.Add("")
	f.Add("hello")
	f.Add("中文")
	f.Add("😀")
	f.Add("\x1b[2mfoo\x1b[0m")
	f.Add("\x1b[38;2;255;0;128mtext\x1b[0m")
	f.Add("e\u0301") // combining accent

	f.Fuzz(func(t *testing.T, s string) {
		got := visibleWidth(s)
		if got < 0 {
			t.Errorf("visibleWidth(%q) = %d, want >= 0", s, got)
		}
	})
}

// FuzzPadRight ensures padRight never panics, returns valid UTF-8, and the
// visible width of the result is exactly `width` (for plain strings) or <= sw
// for ANSI strings when sw > width.
func FuzzPadRight(f *testing.F) {
	f.Add("", 0)
	f.Add("hello", 10)
	f.Add("中文", 5)
	f.Add("hello world extra long string", 10)
	f.Add("中文中文中文", 8)
	f.Add("\x1b[2mfoo\x1b[0m", 5)

	f.Fuzz(func(t *testing.T, s string, width int) {
		if width < 0 || width > 1000 {
			return // avoid extreme widths slowing the fuzzer
		}
		if !utf8.ValidString(s) {
			return // only fuzz valid UTF-8 inputs
		}
		out := padRight(s, width)
		if !utf8.ValidString(out) {
			t.Errorf("padRight(%q, %d) = %q: not valid UTF-8", s, width, out)
		}
		w := visibleWidth(out)
		if w < 0 {
			t.Errorf("padRight(%q, %d): visibleWidth = %d < 0", s, width, w)
		}
		// For plain strings (no ANSI), result visible width must equal width
		// when the input fits, or be <= width when truncated.
		if !ansiEscapeRe.MatchString(s) && w > width {
			t.Errorf("padRight(%q, %d): visible width %d > %d", s, width, w, width)
		}
	})
}

// FuzzFormatKeyFull ensures formatKeyFull never panics.
// Note: the function byte-slices at len(key)-8, so for non-ASCII keys
// the output may contain invalid UTF-8 — the fuzz test only guards against panics.
func FuzzFormatKeyFull(f *testing.F) {
	f.Add("")
	f.Add("short")
	f.Add("12345678")
	f.Add("123456789")
	f.Add("npub1" + "abcdefghijklmnopqrstuvwxyz234567" + "abcdefghijklmnopqrstuvwxyz234567")

	f.Fuzz(func(t *testing.T, key string) {
		// Must not panic.
		_ = formatKeyFull(key)
	})
}

// FuzzPubkeyColor ensures pubkeyColor never panics for any input.
func FuzzPubkeyColor(f *testing.F) {
	f.Add("")
	f.Add("03d884d6abcdef1234567890")
	f.Add("a")
	f.Add("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")

	f.Fuzz(func(t *testing.T, key string) {
		r, g, b := pubkeyColor(key)
		_ = r
		_ = g
		_ = b
	})
}

// FuzzChatBackspace ensures the chat model's inputText remains valid UTF-8
// after arbitrary typing followed by repeated backspaces.
func FuzzChatBackspace(f *testing.F) {
	f.Add("hello")
	f.Add("中文")
	f.Add("😀test")
	f.Add("éàü")
	f.Add("hello中world")

	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			return // only fuzz valid UTF-8 inputs
		}

		m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
		m.typing = true

		// Type the string character by character.
		for _, r := range s {
			m, _ = m.update(pressRealChar(r, string(r)))
		}

		if !utf8.ValidString(m.inputText) {
			t.Fatalf("after typing %q: inputText %q is not valid UTF-8", s, m.inputText)
		}

		// Backspace until empty — inputText must always be valid UTF-8.
		for m.inputText != "" {
			m, _ = m.update(tea.KeyPressMsg{Code: tea.KeyBackspace})
			if !utf8.ValidString(m.inputText) {
				t.Fatalf("during backspace of %q: inputText %q is not valid UTF-8", s, m.inputText)
			}
		}
	})
}

// FuzzRoomLine ensures roomLine never panics and returns valid UTF-8.
func FuzzRoomLine(f *testing.F) {
	f.Add("srv", "room", false)
	f.Add("中文", "房間", false)
	f.Add("", "", true)
	f.Add("very-long-server-name-here", "very-long-room-name-here-too", true)

	f.Fuzz(func(t *testing.T, serverName, roomName string, hasPassword bool) {
		if !utf8.ValidString(serverName) || !utf8.ValidString(roomName) {
			return
		}
		srv := serverConfig{Quickname: serverName}
		room := generated.Room{Name: &roomName, HasPassword: &hasPassword}
		sr := serverRoom{server: srv, room: room}
		line := roomLine(sr, "  ")
		if !utf8.ValidString(line) {
			t.Errorf("roomLine: result is not valid UTF-8: %q", line)
		}
	})
}
