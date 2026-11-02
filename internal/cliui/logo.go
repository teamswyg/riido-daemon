package cliui

import (
	"fmt"
	"strings"
)

// LogoWidth is the number of cells the full logo needs.
func LogoWidth() int { return rowsWidth(logoRows) }

// SymbolWidth is the number of cells the round symbol needs.
func SymbolWidth() int { return rowsWidth(symbolRows) }

func rowsWidth(rows [][][2]uint32) int {
	width := 0
	for _, row := range rows {
		width = max(width, len(row))
	}
	return width
}

// Logo renders the Riido signature with half-block characters. A terminal
// too narrow for the signature gets only the round symbol. It returns "" when
// neither fits or there is no color, so callers fall back to a text title.
func (t Terminal) Logo() string {
	if !t.Interactive || !t.Color {
		return ""
	}
	switch {
	case t.Width >= LogoWidth()+2:
		return t.renderCells(logoRows)
	case t.Width >= SymbolWidth()+2:
		return t.renderCells(symbolRows)
	}
	return ""
}

func (t Terminal) renderCells(rows [][][2]uint32) string {
	var b strings.Builder
	for _, row := range rows {
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
