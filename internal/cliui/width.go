package cliui

import "strings"

// DisplayWidth counts terminal cells: East Asian wide characters (Hangul,
// CJK, fullwidth forms, most emoji) take two cells.
func DisplayWidth(s string) int {
	width := 0
	for _, r := range StripANSI(s) {
		switch {
		case r == 0 || r < 0x20 || (r >= 0x7f && r < 0xa0) || (r >= 0x300 && r <= 0x36f) || r == 0x200b:
		case wide(r):
			width += 2
		default:
			width++
		}
	}
	return width
}

func wide(r rune) bool {
	return (r >= 0x1100 && r <= 0x115f) || (r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
		(r >= 0xac00 && r <= 0xd7a3) || (r >= 0xf900 && r <= 0xfaff) || (r >= 0xfe30 && r <= 0xfe4f) ||
		(r >= 0xff00 && r <= 0xff60) || (r >= 0xffe0 && r <= 0xffe6) || (r >= 0x1f300 && r <= 0x1f64f) ||
		(r >= 0x1f900 && r <= 0x1f9ff) || (r >= 0x20000 && r <= 0x3fffd)
}

// PadRight pads s with spaces to width cells.
func PadRight(s string, width int) string {
	if gap := width - DisplayWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

// Truncate shortens plain text to at most width cells, ending with "…".
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if DisplayWidth(s) <= width {
		return s
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		w := DisplayWidth(string(r))
		if used+w > width-1 {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String() + "…"
}
