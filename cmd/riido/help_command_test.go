package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/cliui"
)

// Curated help must not invent commands or flags: each one has to exist in
// the complete legacy usage that `riido help --all` prints.
func TestCommandHelpCatalogMatchesLegacyUsage(t *testing.T) {
	var legacy bytes.Buffer
	printCommandUsage(&legacy, false)
	usage := legacy.String()
	public := map[string]bool{}
	for _, group := range friendlyHelpGroups {
		for _, entry := range group.entries {
			public[entry.name] = true
		}
	}
	for name, topic := range commandHelpCatalog {
		if topic.name != name || !public[name] {
			t.Fatalf("catalog entry %q is not a public command", name)
		}
		for _, sub := range topic.subs {
			if !strings.Contains(usage, "riido "+name+" "+sub.name) && !strings.Contains(usage, sub.name+"|") && !strings.Contains(usage, "|"+sub.name) {
				t.Fatalf("riido %s %s is not in the legacy usage", name, sub.name)
			}
			for _, flag := range sub.flags {
				requireLegacyFlag(t, usage, name, flag)
			}
		}
		for _, flag := range topic.flags {
			requireLegacyFlag(t, usage, name, flag)
		}
	}
}

func requireLegacyFlag(t *testing.T, usage, command string, flag helpEntry) {
	t.Helper()
	token := strings.Fields(flag.name)[0]
	if !strings.Contains(usage, token) {
		t.Fatalf("%s flag %s is not in the legacy usage", command, token)
	}
}

func TestCommandHelpFollowsGitHubCLILayout(t *testing.T) {
	var buf bytes.Buffer
	printCommandHelp(&buf, cliui.Plain, commandHelpCatalog["work"])
	out := buf.String()
	order := []string{"작업과 댓글", "\n사용법\n", "\n명령\n", "  get ", "\n옵션\n", "--workspace <ID>", "\n예시\n", "$ riido work get RIID-1234", "\n더 알아보기\n"}
	last := -1
	for _, part := range order {
		i := strings.Index(out, part)
		if i <= last {
			t.Fatalf("%q missing or out of order:\n%s", part, out)
		}
		last = i
	}
	buf.Reset()
	work := commandHelpCatalog["work"]
	printSubcommandHelp(&buf, cliui.Plain, work, *work.sub("search"))
	if !strings.Contains(buf.String(), "riido work search --query <TEXT> [옵션]") || !strings.Contains(buf.String(), "--query <TEXT>") || !strings.Contains(buf.String(), "--team-id <ID>") {
		t.Fatalf("subcommand help:\n%s", buf.String())
	}
}

func TestSuggestCommandFindsCloseNames(t *testing.T) {
	names := []string{"auth", "work", "workspace", "meeting", "doctor"}
	for input, want := range map[string]string{"wrk": "work", "meetin": "meeting", "docter": "doctor", "wor": "work", "zzzzzz": "", "a": ""} {
		if got := suggestCommand(input, names); got != want {
			t.Fatalf("suggest(%q) = %q, want %q", input, got, want)
		}
	}
	if !looksLikeWorkKey("RIID-1234") || looksLikeWorkKey("get") || looksLikeWorkKey("RIID-") {
		t.Fatal("work key detection")
	}
}

func TestRootHelpShowsExamplesAndLearnMore(t *testing.T) {
	var buf bytes.Buffer
	printFriendlyHelp(&buf, cliui.Plain)
	for _, part := range []string{"예시", "$ riido work get RIID-1234", "더 알아보기", "riido help --all"} {
		if !strings.Contains(buf.String(), part) {
			t.Fatalf("root help missing %q:\n%s", part, buf.String())
		}
	}
}
