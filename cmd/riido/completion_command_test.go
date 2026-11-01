package main

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/clicompletion"
)

func TestCompletionRootHelp(t *testing.T) {
	if err := run([]string{"completion", "--help"}); err != nil {
		t.Fatal(err)
	}
}

func TestCompletionNeverOpensConfigurationOrEnvironment(t *testing.T) {
	for _, shell := range completionShells {
		h := newContextHarness(t, nil)
		h.io.getenv = func(string) string { t.Fatal("completion read environment"); return "" }
		h.io.configDir = func() (string, error) { t.Fatal("completion opened settings"); return "", nil }
		if err := runCompletion(h.io, []string{shell, "--json"}); err != nil {
			t.Fatal(err)
		}
		data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
		if data["shell"] != shell || !strings.Contains(data["script"].(string), "completion __complete") || len(h.daemon.requests) != 0 {
			t.Fatal("wrong script or IPC")
		}
	}
	h := newContextHarness(t, nil)
	h.io.stdin = strings.NewReader("words\x00config\x00list\x00--profile\x00st\x00")
	h.io.getenv = func(string) string { t.Fatal("query read environment"); return "" }
	if err := runCompletion(h.io, []string{"__complete"}); err != nil || h.stdout.String() != "staging\n" || h.stderr.Len() != 0 {
		t.Fatal(err, h.stdout.String(), h.stderr.String())
	}
}

func TestCompletionOpaqueAndSecretInputsAreSilent(t *testing.T) {
	for _, raw := range []string{"words\x00work\x00get\x00--team-id\x00synthetic-secret\x00", "words\x00auth\x00login\x00--token\x00synthetic-secret\x00", "words\x00daemon\x00\x00", "words\x00work\x00search\x00--query\x00--token\x00--format\x00j\x00", "words\x00work\x00search\x00--query\x00--password=synthetic-secret\x00--format\x00j\x00", "malformed synthetic-secret", strings.Repeat("s", clicompletion.MaxInput+1)} {
		h := newContextHarness(t, nil)
		h.io.stdin = strings.NewReader(raw)
		if err := runCompletion(h.io, []string{"__complete"}); err != nil || h.stdout.Len() != 0 || h.stderr.Len() != 0 || len(h.daemon.requests) != 0 {
			t.Fatal("private input returned output")
		}
	}
}

func TestCompletionPublicCatalogMatchesSSOT(t *testing.T) {
	raw, err := os.ReadFile("../../docs/30-architecture/cli-surface.riido.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Groups []struct {
			Name        string
			RootHelp    string `json:"root_help"`
			Subcommands []string
		} `json:"command_groups"`
	}
	if json.Unmarshal(raw, &doc) != nil {
		t.Fatal("invalid manifest")
	}
	roots := clicompletion.Candidates(completionCatalog(), []string{""})
	for _, group := range doc.Groups {
		if slices.Contains(roots, group.Name) != (group.RootHelp == "public") {
			t.Fatal("public root drift", group.Name)
		}
		if group.RootHelp != "public" {
			continue
		}
		children := clicompletion.Candidates(completionCatalog(), []string{group.Name, ""})
		for _, sub := range group.Subcommands {
			unsupportedUser := group.Name == "context" && (sub == "part" || sub == "changes")
			if slices.Contains(children, sub) == unsupportedUser {
				t.Fatal("subcommand drift", group.Name, sub)
			}
		}
	}
}

func TestCompletionDoesNotRecommendUnsupportedUserOptions(t *testing.T) {
	for _, tc := range []struct {
		words  []string
		absent []string
	}{
		{[]string{"auth", "login", "--"}, []string{"--desktop", "--with-token", "--workspace", "--team", "--token"}},
		{[]string{"work", "tree", "--"}, []string{"--page", "--limit", "--cursor", "--budget"}},
		{[]string{"meeting", "list", "--"}, []string{"--page", "--limit", "--cursor", "--budget"}},
		{[]string{"thread", "list", "--"}, []string{"--cursor", "--page", "--budget", "--team"}},
		{[]string{"work", "comments", "--"}, []string{"--comment-id"}},
		{[]string{"work", "body", "--"}, []string{"--board-cursor", "--board-limit"}},
		{[]string{"serve", "--"}, []string{"--profile", "--json", "--workspace"}},
	} {
		choices := clicompletion.Candidates(completionCatalog(), tc.words)
		for _, bad := range tc.absent {
			if slices.Contains(choices, bad) {
				t.Fatal(tc.words, bad)
			}
		}
	}
}
