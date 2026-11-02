package main

import (
	"io"
	"os"
	"slices"

	"github.com/teamswyg/riido-daemon/internal/cliui"
	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

// noColorRequested records a --no-color flag anywhere on the command line.
var noColorRequested bool

func noteNoColor(args []string) { noColorRequested = slices.Contains(args, "--no-color") }

// humanTerminal returns the styling allowed for w. Agent runs, pipes, files,
// test buffers and RIIDO_CLI_PLAIN=1 always get plain output, so scripts and
// agents see exactly the previous text.
func humanTerminal(w io.Writer) cliui.Terminal {
	if contextcli.DetectEnvironment(os.Getenv).Mode == contextcli.ModeAgent {
		return cliui.Plain
	}
	term := cliui.Detect(w, os.Getenv)
	if noColorRequested {
		term.Color, term.TrueColor = false, false
	}
	return term
}
