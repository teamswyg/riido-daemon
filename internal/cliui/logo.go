package cliui

import (
	"fmt"
	"strings"
)

// LogoWidth is the number of cells the full logo needs.
func LogoWidth() int {
	width := 0
	for _, row := range logoRows {
		width = max(width, len(row))
	}
	return width
}

// Logo renders the Riido signature with half-block characters. It returns ""
// when the terminal cannot show it well (no color or too narrow), so callers
// fall back to a text title.
func (t Terminal) Logo() string {
	if !t.Interactive || !t.Color || t.Width < LogoWidth()+2 {
		return ""
	}
	var b strings.Builder
	for _, row := range logoRows {
		for _, cell := range row {
			top, bottom := cell[0], cell[1]
			switch {
			case top != 0 && bottom != 0:
				fmt.Fprintf(&b, "\x1b[38;%s;48;%sm▀\x1b[0m", t.rgb(top&0xffffff), t.rgb(bottom&0xffffff))
			case top != 0:
				fmt.Fprintf(&b, "\x1b[38;%sm▀\x1b[0m", t.rgb(top&0xffffff))
			case bottom != 0:
				fmt.Fprintf(&b, "\x1b[38;%sm▄\x1b[0m", t.rgb(bottom&0xffffff))
			default:
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

// Mark is the one-line fallback title.
func (t Terminal) Mark() string {
	if !t.Interactive {
		return "Riido"
	}
	return t.Accent("◆") + " " + t.Bold("Riido")
}
