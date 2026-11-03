package cliui

import "strings"

// SideBySide places right next to left, vertically centered on it. Every left
// line must already be leftWidth cells wide (padding is added for rows the
// left block does not cover). Trailing spaces are trimmed.
func SideBySide(left []string, leftWidth, gap int, right []string) []string {
	rows := max(len(left), len(right))
	offset := max((len(left)-len(right))/2, 0)
	blank := strings.Repeat(" ", leftWidth)
	out := make([]string, 0, rows)
	for i := range rows {
		line := blank
		if i < len(left) {
			line = left[i]
		}
		if j := i - offset; j >= 0 && j < len(right) && right[j] != "" {
			line += strings.Repeat(" ", gap) + right[j]
		}
		out = append(out, strings.TrimRight(line, " "))
	}
	return out
}
