package tui

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"github.com/EwenQuim/microchat/client/sdk/generated"
	"github.com/mattn/go-runewidth"
)

// ── visibleWidth ────────────────────────────────────────────────────────────

func TestVisibleWidth_UnicodeEdgeCases(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"ASCII", "hello", 5},
		{"CJK double-width", "中文", 4},
		{"ANSI-wrapped ASCII", "\x1b[2mfoo\x1b[0m", 3},
		{"ANSI-wrapped CJK", "\x1b[2m中文\x1b[0m", 4},
		{"combining char (e + acute)", "e\u0301", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := visibleWidth(tc.input)
			if got < 0 {
				t.Errorf("visibleWidth(%q) = %d, want >= 0", tc.input, got)
			}
			if got != tc.want {
				t.Errorf("visibleWidth(%q) = %d, want %d", tc.input, got, tc.want)
			}
		})
	}
}

// ── padRight ────────────────────────────────────────────────────────────────

func TestPadRight_WidthBehavior(t *testing.T) {
	t.Run("pads short string to width", func(t *testing.T) {
		s := padRight("hi", 10)
		if w := runewidth.StringWidth(s); w != 10 {
			t.Errorf("padRight(\"hi\", 10): width = %d, want 10", w)
		}
	})

	t.Run("exact-width string unchanged", func(t *testing.T) {
		s := padRight("hello", 5)
		if s != "hello" {
			t.Errorf("padRight(\"hello\", 5) = %q, want %q", s, "hello")
		}
	})

	t.Run("over-wide ASCII truncated to width", func(t *testing.T) {
		long := strings.Repeat("a", 20)
		s := padRight(long, 10)
		w := visibleWidth(s)
		if w > 10 {
			t.Errorf("padRight over-wide ASCII: width = %d, want <= 10", w)
		}
		if !utf8.ValidString(s) {
			t.Errorf("padRight over-wide ASCII: result is not valid UTF-8: %q", s)
		}
	})

	t.Run("over-wide CJK truncated to width", func(t *testing.T) {
		s := padRight("中文中文中文中文", 10)
		w := visibleWidth(s)
		if w > 10 {
			t.Errorf("padRight over-wide CJK: width = %d, want <= 10", w)
		}
		if !utf8.ValidString(s) {
			t.Errorf("padRight over-wide CJK: result is not valid UTF-8: %q", s)
		}
	})

	t.Run("zero width does not panic", func(t *testing.T) {
		_ = padRight("hi", 0)
	})
}

// ── formatKeyFull ────────────────────────────────────────────────────────────

func TestFormatKeyFull_UTF8Safety(t *testing.T) {
	// formatKeyFull is used with npub keys (ASCII bech32). Test that pure ASCII
	// keys are always handled safely.
	cases := []struct {
		key string
	}{
		{""},
		{"short"},
		{"12345678"},
		{"123456789"},
		{strings.Repeat("a", 64)},
		{"npub1" + strings.Repeat("z", 59)}, // typical npub length
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("len=%d", len(tc.key)), func(t *testing.T) {
			out := formatKeyFull(tc.key)
			if !utf8.ValidString(out) {
				t.Errorf("formatKeyFull(%q) = %q: not valid UTF-8", tc.key, out)
			}
		})
	}
}

// ── pubkeyColor ──────────────────────────────────────────────────────────────

func TestPubkeyColor_AlwaysValidRGB(t *testing.T) {
	inputs := []string{
		"",
		"a",
		"03d884d6abcdef1234567890",
		strings.Repeat("x", 100),
		strings.Repeat("y", 200),
	}
	for _, k := range inputs {
		t.Run(fmt.Sprintf("key=%q", k[:min(len(k), 16)]), func(t *testing.T) {
			// uint8 values are always in [0,255] — just ensure no panic.
			r, g, b := pubkeyColor(k)
			_ = r
			_ = g
			_ = b
		})
	}
}

func TestPubkeyColor_HueSpread(t *testing.T) {
	distinct := make(map[[3]uint8]struct{})
	for i := range 50 {
		key := fmt.Sprintf("pubkey%050d", i)
		r, g, b := pubkeyColor(key)
		distinct[[3]uint8{r, g, b}] = struct{}{}
	}
	if len(distinct) < 10 {
		t.Errorf("hue spread: only %d distinct colors for 50 keys, expected >= 10", len(distinct))
	}
}

// ── chat backspace (Bug 1: byte-slices multi-byte runes) ─────────────────────

func TestChatBackspace_MultiByteUTF8(t *testing.T) {
	cases := []struct {
		name  string
		input string // a single character (or multi-char string ending in the char to delete)
		want  string // expected inputText after one backspace
	}{
		{"CJK 3 bytes", "中", ""},
		{"emoji 4 bytes", "😀", ""},
		{"accented Latin 2 bytes", "é", ""},
		{"mixed: ASCII then CJK", "hello中", "hello"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
			m.typing = true

			// Type the string character-by-character.
			for _, r := range tc.input {
				m, _ = m.update(pressRealChar(r, string(r)))
			}
			if m.inputText != tc.input {
				t.Fatalf("after typing: inputText = %q, want %q", m.inputText, tc.input)
			}

			// Press backspace once.
			m, _ = m.update(pressKey(tea.KeyBackspace))

			if !utf8.ValidString(m.inputText) {
				t.Errorf("after backspace: inputText %q is not valid UTF-8", m.inputText)
			}
			if m.inputText != tc.want {
				t.Errorf("after backspace: inputText = %q, want %q", m.inputText, tc.want)
			}
		})
	}
}

// ── non-chat inputs accept multi-byte chars (Bug 3: len(s)==1 drops runes) ───

func TestNonChatInput_AcceptsMultiByteChars(t *testing.T) {
	t.Run("room search", func(t *testing.T) {
		m := newRoomModel(nil, nil)
		m.state = roomStateSearch
		m, _ = m.update(pressChar("中"))
		if m.inputText != "中" {
			t.Errorf("inputText = %q, want %q", m.inputText, "中")
		}
	})

	t.Run("room create", func(t *testing.T) {
		m := newRoomModel(nil, nil)
		m.state = roomStateCreate
		m, _ = m.update(pressChar("中"))
		if m.inputText != "中" {
			t.Errorf("inputText = %q, want %q", m.inputText, "中")
		}
	})

	t.Run("room passwd", func(t *testing.T) {
		m := newRoomModel(nil, nil)
		m.promptPasswd = true
		m, _ = m.update(pressChar("中"))
		if m.passwdInput != "中" {
			t.Errorf("passwdInput = %q, want %q", m.passwdInput, "中")
		}
	})

	t.Run("server URL", func(t *testing.T) {
		m := serverModel{state: serverStateAddURL}
		m2, _ := m.update(pressChar("é"))
		if m2.inputText != "é" {
			t.Errorf("inputText = %q, want %q", m2.inputText, "é")
		}
	})

	t.Run("contact npub", func(t *testing.T) {
		m := contactsModel{state: contactsStateAddNpub}
		m2, _ := m.update(pressChar("中"))
		if m2.inputNpub != "中" {
			t.Errorf("inputNpub = %q, want %q", m2.inputNpub, "中")
		}
	})

	t.Run("contact name", func(t *testing.T) {
		m := contactsModel{state: contactsStateAddName}
		m2, _ := m.update(pressChar("中"))
		if m2.inputName != "中" {
			t.Errorf("inputName = %q, want %q", m2.inputName, "中")
		}
	})
}

// ── non-chat backspace on multi-byte chars (Bug 3 + backspace fix) ────────────

func TestNonChatBackspace_MultiByteClean(t *testing.T) {
	assertCleanBackspace := func(t *testing.T, label, got string) {
		t.Helper()
		if !utf8.ValidString(got) {
			t.Errorf("%s: after backspace, result %q is not valid UTF-8", label, got)
		}
		if got != "" {
			t.Errorf("%s: after backspace on single rune, result = %q, want %q", label, got, "")
		}
	}

	t.Run("room search", func(t *testing.T) {
		m := newRoomModel(nil, nil)
		m.state = roomStateSearch
		m.inputText = "中"
		m, _ = m.update(pressKey(tea.KeyBackspace))
		assertCleanBackspace(t, "room search", m.inputText)
	})

	t.Run("room create", func(t *testing.T) {
		m := newRoomModel(nil, nil)
		m.state = roomStateCreate
		m.inputText = "中"
		m, _ = m.update(pressKey(tea.KeyBackspace))
		assertCleanBackspace(t, "room create", m.inputText)
	})

	t.Run("room passwd", func(t *testing.T) {
		m := newRoomModel(nil, nil)
		m.promptPasswd = true
		m.passwdInput = "中"
		m, _ = m.update(pressKey(tea.KeyBackspace))
		assertCleanBackspace(t, "room passwd", m.passwdInput)
	})

	t.Run("server URL", func(t *testing.T) {
		m := serverModel{state: serverStateAddURL, inputText: "中"}
		m2, _ := m.update(pressKey(tea.KeyBackspace))
		assertCleanBackspace(t, "server URL", m2.inputText)
	})

	t.Run("contact npub", func(t *testing.T) {
		m := contactsModel{state: contactsStateAddNpub, inputNpub: "中"}
		m2, _ := m.update(pressKey(tea.KeyBackspace))
		assertCleanBackspace(t, "contact npub", m2.inputNpub)
	})

	t.Run("contact name", func(t *testing.T) {
		m := contactsModel{state: contactsStateAddName, inputName: "中"}
		m2, _ := m.update(pressKey(tea.KeyBackspace))
		assertCleanBackspace(t, "contact name", m2.inputName)
	})
}

// ── roomLine with long / CJK names ───────────────────────────────────────────

func TestRoomLine_LongNames(t *testing.T) {
	cases := []struct {
		name       string
		serverName string
		roomName   string
	}{
		{"long server name", strings.Repeat("a", 30), "room"},
		{"long room name", "srv", strings.Repeat("b", 40)},
		{"all-CJK server", "中文中文中文中", "房間"},
		{"all-CJK room", "srv", "房間房間房間房間房間房間"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := serverConfig{Quickname: tc.serverName}
			roomName := tc.roomName
			sr := serverRoom{server: srv, room: generated.Room{Name: &roomName}}
			line := roomLine(sr, "  ")
			if !utf8.ValidString(line) {
				t.Errorf("roomLine: result is not valid UTF-8: %q", line)
			}
			// After truncation, the visible content should fit within column bounds.
			// Strip trailing newline for width measurement.
			trimmed := strings.TrimSuffix(line, "\n")
			w := visibleWidth(trimmed)
			maxW := 1 + 2 + maxServerNameWidth + 1 + maxRoomNameWidth + 4 // prefix + cursor + server~ + room + lock + padding
			if w > maxW {
				t.Errorf("roomLine: visible width %d > max %d for line %q", w, maxW, line)
			}
		})
	}
}

// ── viewPanel degenerate terminal sizes ──────────────────────────────────────

func TestViewPanel_DoesNotPanic(t *testing.T) {
	m := newChatModel(nil, serverConfig{}, "room", "", nil, "alice")
	m.loading = false

	for _, size := range [][2]int{{0, 0}, {1, 1}, {10000, 1}, {80, 24}} {
		w, h := size[0], size[1]
		t.Run(fmt.Sprintf("%dx%d", w, h), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("viewPanel(%d, %d) panicked: %v", w, h, r)
				}
			}()
			_ = m.viewPanel(w, h, true)
		})
	}
}
