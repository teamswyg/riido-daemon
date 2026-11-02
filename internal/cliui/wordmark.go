package cliui

import "strings"

// wordmarkPattern is the "Riido" name drawn by hand on the same 14-pixel grid
// as the round symbol ('#' is a pixel). Scaling the signature artwork down to
// terminal resolution broke the letters apart and left stray pixels, so the
// letters are drawn on whole pixels with two-pixel strokes instead.
var wordmarkPattern = []string{
	".......................................",
	"#######....##..##.........##...........",
	"########...##..##.........##...........",
	"##....##..................##...........",
	"##....##..................##...........",
	"##....##...##..##....#######....#####..",
	"########...##..##...########...#######.",
	"#######....##..##..##.....##..##.....##",
	"##..##.....##..##..##.....##..##.....##",
	"##...##....##..##..##.....##..##.....##",
	"##....##...##..##..##.....##..##.....##",
	"##.....##..##..##...########...#######.",
	"##.....##..##..##....#######....#####..",
	".......................................",
}

// logoGap is the space between the symbol and the wordmark, in cells.
const logoGap = 3

// LogoWidth is the number of cells the symbol plus the wordmark need.
func LogoWidth() int { return SymbolWidth() + logoGap + len(wordmarkPattern[0]) }

// LogoLines renders the round symbol with the pixel "Riido" wordmark beside
// it, one string per terminal row. The wordmark uses the accent purple chosen
// for the terminal background. It is nil without color or when the terminal
// is narrower than the logo, so callers fall back to the symbol alone.
func (t Terminal) LogoLines() []string {
	symbol := t.SymbolLines()
	if symbol == nil || t.Width < LogoWidth()+2 || len(wordmarkPattern) != 2*len(symbol) {
		return nil
	}
	color := t.accentRGB() | 1<<24
	lines := make([]string, len(symbol))
	for row := range symbol {
		var b strings.Builder
		b.WriteString(symbol[row])
		b.WriteString(strings.Repeat(" ", logoGap))
		top, bottom := wordmarkPattern[2*row], wordmarkPattern[2*row+1]
		for i := range len(top) {
			var upper, lower uint32
			if top[i] == '#' {
				upper = color
			}
			if bottom[i] == '#' {
				lower = color
			}
			t.writeCell(&b, upper, lower)
		}
		lines[row] = strings.TrimRight(b.String(), " ")
	}
	return lines
}
