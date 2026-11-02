package cliui

import (
	"io"
	"strings"
)

// Table writes aligned columns. Interactive terminals get a dim header and the
// last column is shortened to the terminal width; plain output keeps every
// value whole and separates columns with two spaces.
func (t Terminal) Table(w io.Writer, headers []string, rows [][]string) error {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = DisplayWidth(h)
	}
	for _, row := range rows {
		for i := range min(len(row), len(widths)) {
			widths[i] = max(widths[i], DisplayWidth(row[i]))
		}
	}
	line := func(cells []string, header bool) string {
		parts := make([]string, len(widths))
		used := 0
		for i := range widths {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			if i == len(widths)-1 {
				if t.Interactive {
					cell = Truncate(cell, max(t.Width-used, 8))
				}
				parts[i] = cell
			} else {
				parts[i] = PadRight(cell, widths[i])
				used += widths[i] + 2
			}
			if header {
				parts[i] = t.Dim(parts[i])
			}
		}
		return strings.TrimRight(strings.Join(parts, "  "), " ") + "\n"
	}
	if _, err := io.WriteString(w, line(headers, true)); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := io.WriteString(w, line(row, false)); err != nil {
			return err
		}
	}
	return nil
}
