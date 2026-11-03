package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliui"
)

func TestFriendlyHelpListsEveryPublicCommandOnce(t *testing.T) {
	var buf bytes.Buffer
	printFriendlyHelp(&buf, cliui.Plain)
	out := buf.String()
	for _, command := range []mainCommand{mainCommandAuth, mainCommandWhoami, mainCommandDoctor, mainCommandWork, mainCommandContext, mainCommandMeeting, mainCommandAgents, mainCommandThread, mainCommandResource, mainCommandRepo, mainCommandWorkspace, mainCommandConfig, mainCommandCompletion, mainCommandCapabilities, mainCommandVersion} {
		if strings.Count(out, "\n  "+string(command)+" ") != 1 {
			t.Fatalf("%s missing or repeated:\n%s", command, out)
		}
	}
	for _, internal := range []string{"serve", "api ", "task ", "daemon", "mwsd", "bridge", "claude-permission-mcp"} {
		if strings.Contains(out, "\n  "+internal) {
			t.Fatalf("internal command %q in friendly help", internal)
		}
	}
	if strings.Contains(out, "\x1b") {
		t.Fatal("plain help styled")
	}
}

func TestHomeFallsBackToTextTitleWithoutColor(t *testing.T) {
	var buf bytes.Buffer
	if err := runHome(&buf, cliui.Terminal{Interactive: true, Width: 120}, "https://monolith.api.riido.io (staging)"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.ContainsAny(out, "▀▄\x1b") || !strings.Contains(out, "Riido CLI") || !strings.Contains(out, "riido auth login") || !strings.Contains(out, "서버 https://monolith.api.riido.io (staging)") {
		t.Fatalf("home without color:\n%s", out)
	}
	buf.Reset()
	_ = runHome(&buf, cliui.Terminal{Interactive: true, Color: true, TrueColor: true, Width: 120}, "")
	if !strings.Contains(buf.String(), "▀") {
		t.Fatal("logo missing on a wide color terminal")
	}
}

func TestHelpAllKeepsCompleteLegacyUsage(t *testing.T) {
	var all, legacy bytes.Buffer
	if err := runHelp(&all, []string{"--all"}); err != nil {
		t.Fatal(err)
	}
	printCommandUsage(&legacy, false)
	if all.String() != legacy.String() {
		t.Fatal("help --all differs from the legacy root usage")
	}
	// A non-terminal writer keeps the legacy text for agents and scripts.
	var piped bytes.Buffer
	if err := runHelp(&piped, nil); err != nil || piped.String() != legacy.String() {
		t.Fatal("piped help changed")
	}
	if err := runHelp(&piped, []string{"--bogus"}); err == nil {
		t.Fatal("unknown help argument accepted")
	}
}

// A narrow terminal stacks each summary under its command instead of wrapping.
func TestHomeStacksQuickStartOnNarrowTerminal(t *testing.T) {
	var buf bytes.Buffer
	if err := runHome(&buf, cliui.Terminal{Interactive: true, Width: 30}, ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "  riido auth login\n    브라우저로 로그인\n") {
		t.Fatalf("narrow quick start not stacked:\n%s", buf.String())
	}
	buf.Reset()
	_ = runHome(&buf, cliui.Terminal{Interactive: true, Width: 120}, "")
	if !strings.Contains(buf.String(), "riido auth login") || strings.Contains(buf.String(), "riido auth login\n") {
		t.Fatalf("wide quick start should keep two columns:\n%s", buf.String())
	}
}

// The home screen leaves a blank line after the prompt and before the next one.
func TestHomeHasBlankLinesAroundIt(t *testing.T) {
	var buf bytes.Buffer
	if err := runHome(&buf, cliui.Terminal{Interactive: true, Color: true, TrueColor: true, Width: 120}, "https://monolith.api.riido.io (staging)"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "\n") || !strings.HasSuffix(out, "\n\n") || !strings.Contains(out, "▀") {
		t.Fatalf("home spacing:\n%q", out)
	}
}

// A wide color terminal shows the symbol with the pixel wordmark, then the
// name and server as text; a medium one keeps the text beside the symbol.
func TestHomePicksLogoByWidth(t *testing.T) {
	logoWidth := func(width int) int {
		var buf bytes.Buffer
		_ = runHome(&buf, cliui.Terminal{Interactive: true, Color: true, TrueColor: true, Width: width}, "")
		widest := 0
		for _, line := range strings.Split(buf.String(), "\n")[1:8] {
			widest = max(widest, cliui.DisplayWidth(line))
		}
		return widest
	}
	if got := logoWidth(120); got <= cliui.SymbolWidth()+3 {
		t.Fatalf("wordmark missing on a wide terminal: %d cells", got)
	}
	if got := logoWidth(cliui.LogoWidth() - 1); got != cliui.SymbolWidth() {
		t.Fatalf("medium terminal should show only the symbol rows: %d cells", got)
	}
}
