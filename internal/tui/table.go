package tui

import (
	"strings"

	table "charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

func tableStyles() table.Styles {
	return table.Styles{
		Header:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
		Selected: lipgloss.NewStyle().Bold(true),
	}
}

func renderTable(cols []table.Column, rows []table.Row, cursor int, pad string) string {
	// Compute total width: each non-zero column occupies Width + 2 (padding 0,1).
	totalWidth := 0
	for _, col := range cols {
		if col.Width > 0 {
			totalWidth += col.Width + 2
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(len(rows)+1),
		table.WithWidth(totalWidth),
	)
	t.SetCursor(cursor)
	t.SetStyles(tableStyles())

	output := t.View()
	lines := strings.Split(output, "\n")
	var b strings.Builder
	for _, line := range lines {
		if line != "" {
			b.WriteString(pad + line + "\n")
		}
	}
	return b.String()
}
