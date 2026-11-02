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
