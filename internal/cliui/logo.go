package cliui

import (
	"fmt"
	"strings"
)

// SymbolWidth is the number of cells the round symbol needs.
func SymbolWidth() int {
	width := 0
	for _, row := range symbolRows {
		width = max(width, len(row))
	}
	return width
}

// SymbolLines renders the round Riido symbol with half-block characters, one
// string per terminal row, each exactly SymbolWidth cells wide. It is nil
// without color or when the terminal is too narrow, so callers show text only.
// The wordmark is never drawn with blocks: at terminal resolution the letters
// break up, so the product name stays real text next to the symbol.
func (t Terminal) SymbolLines() []string {
	if !t.Interactive || !t.Color || t.Width < SymbolWidth()+2 {
		return nil
	}
	width := SymbolWidth()
	lines := make([]string, 0, len(symbolRows))
	for _, row := range symbolRows {
		var b strings.Builder
		for i := range width {
			var cell [2]uint32
			if i < len(row) {
				cell = row[i]
			}
			t.writeCell(&b, cell[0], cell[1])
		}
		lines = append(lines, b.String())
	}
	return lines
}

func (t Terminal) writeCell(b *strings.Builder, top, bottom uint32) {
	switch {
	case top != 0 && top == bottom:
		// A full block avoids the seam some terminals draw between fg and bg.
		fmt.Fprintf(b, "\x1b[38;%sm█\x1b[0m", t.rgb(top&0xffffff))
	case top != 0 && bottom != 0:
		fmt.Fprintf(b, "\x1b[38;%s;48;%sm▀\x1b[0m", t.rgb(top&0xffffff), t.rgb(bottom&0xffffff))
	case top != 0:
		fmt.Fprintf(b, "\x1b[38;%sm▀\x1b[0m", t.rgb(top&0xffffff))
	case bottom != 0:
		fmt.Fprintf(b, "\x1b[38;%sm▄\x1b[0m", t.rgb(bottom&0xffffff))
	default:
		b.WriteByte(' ')
	}
}

// Mark is the one-line title used where the symbol does not fit.
func (t Terminal) Mark() string {
	if !t.Interactive {
		return "Riido"
	}
	return t.Accent("◆") + " " + t.Bold("Riido")
}
