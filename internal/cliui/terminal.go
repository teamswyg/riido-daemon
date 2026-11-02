// Package cliui renders human-facing terminal output for the Riido CLI. It is
// used only when a person is looking at an interactive terminal; pipes, files,
// agent runs and --json keep their original plain output byte for byte.
package cliui

import (
	"io"
	"os"
	"strconv"
	"strings"
)

// Terminal describes what one output stream can show.
type Terminal struct {
	// Interactive is a real terminal with a usable TERM.
	Interactive bool
	// Color allows ANSI styling (Interactive and NO_COLOR unset).
	Color bool
	// TrueColor allows 24-bit colors; otherwise the 256-color palette is used.
	TrueColor bool
	// Width is the terminal width in cells (80 when unknown).
	Width int
}

// Plain is the zero-styling terminal used for pipes and files.
var Plain = Terminal{Width: 80}

// Detect inspects w. Only an *os.File connected to a terminal is interactive.
// RIIDO_CLI_PLAIN=1 or TERM=dumb force plain output; NO_COLOR keeps layout
// but removes color (https://no-color.org).
func Detect(w io.Writer, getenv func(string) string) Terminal {
	if getenv == nil {
		getenv = os.Getenv
	}
	file, ok := w.(*os.File)
	if !ok || file == nil || getenv("RIIDO_CLI_PLAIN") == "1" || strings.EqualFold(getenv("TERM"), "dumb") {
		return Plain
	}
	// A character device is not enough: /dev/null is one too, and a redirect
	// there must keep the non-interactive exit behavior.
	if !isTerminal(file) {
		return Plain
	}
	term := Terminal{Interactive: true, Width: 80}
	if columns, err := strconv.Atoi(getenv("COLUMNS")); err == nil && columns > 0 {
		term.Width = columns
	} else if columns := terminalWidth(file); columns > 0 {
		term.Width = columns
	}
	term.Color = getenv("NO_COLOR") == ""
	colorterm := strings.ToLower(getenv("COLORTERM"))
	term.TrueColor = term.Color && (colorterm == "truecolor" || colorterm == "24bit")
	return term
}
