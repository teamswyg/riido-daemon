package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserAgentsRejectInvalidArgumentsBeforeCredentials(t *testing.T) {
	for _, args := range [][]string{
		{"list"},
		{"show", "a"},
		{"list", "extra", "--work", "t", "--workspace", "w"},
		{"list", "--work", "../t", "--workspace", "w"},
		{"list", "--work", "RIID-5972", "--workspace", "w"},
		{"show", "a/b", "--work", "t", "--workspace", "w"},
		{"show", "a", "--work", "t", "--workspace", "w", "--cursor", "0"},
		{"list", "--work", "t", "--workspace", "w", "--cursor", ""},
		{"list", "--work", "t", "--workspace", "w", "--cursor", "01"},
		{"list", "--work", "t", "--workspace", "w", "--cursor", "1&cursor=2"},
		{"list", "--work", "t", "--workspace", "w", "--limit", "5"},
		{"list", "--work", "t", "--workspace", "w", "--budget", "1"},
		{"list", "--token", "secret"},
		{"list", "--cp-url", "https://other.invalid"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { t.Error("opened credentials"); return nil, nil }
			if err := runAgents(h.io, append(args, "--json")); err == nil {
				t.Fatal("invalid arguments accepted")
			}
		})
	}
	h := newContextHarness(t, agentVars())
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Error("agent opened user session")
		return nil, nil
	}
	if err := runAgents(h.io, []string{"list", "--work", "task-one", "--workspace", "workspace-one", "--json"}); err == nil {
		t.Fatal("agent mode accepted user-only work selector")
	}
}

func TestUserAgentsErrorMappingAndTerminalSafety(t *testing.T) {
	for _, tc := range []struct{ status, code int }{{401, 3}, {403, 4}, {503, 6}, {429, 7}} {
		s, _ := userAgentsCLIService(t, "private-response-canary", tc.status)
		h := newContextHarness(t, nil)
		h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
		if code := exitCodeOf(t, runAgents(h.io, []string{"list", "--work", "task-one", "--workspace", "workspace-one", "--json"})); code != tc.code {
			t.Fatal("wrong stable error", tc, code)
		}
		if strings.Contains(h.stdout.String()+h.stderr.String(), "private-response-canary") {
			t.Fatal("provider body leaked")
		}
	}
	body := `{"agents":[` + strings.Replace(userAgentFixture, "리뷰 Agent", `이름\u001b[31m\n`, 1) + `],"total":1,"cursor":"0","fetched_at":"2026-11-01T01:00:00Z","policy_checked":false}`
	s, _ := userAgentsCLIService(t, body, 200)
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
	if err := runAgents(h.io, []string{"list", "--work", "task-one", "--workspace", "workspace-one"}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(h.stdout.String(), "\x1b") || !strings.Contains(h.stdout.String(), `\x1b`) || !strings.Contains(h.stdout.String(), "멘션·실행 허가") {
		t.Fatal("terminal output lost safety or execution boundary", h.stdout.String())
	}
}

func TestUserAgentsCLIPageAndCapabilities(t *testing.T) {
	entries := []string{}
	for _, id := range []string{"a", "b", "c"} {
		entries = append(entries, strings.Replace(userAgentFixture, "agent-one", id, 1))
	}
	body := `{"agents":[` + strings.Join(entries, ",") + `],"total":53,"cursor":"1","fetched_at":"2026-11-01T01:00:00Z","policy_checked":false}`
	s, calls := userAgentsCLIService(t, body, 200)
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
	if err := runAgents(h.io, []string{"list", "--work", "task-one", "--workspace", "workspace-one", "--cursor", "1", "--json"}); err != nil {
		t.Fatal(err)
	}
	data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)["directory"].(map[string]any)
	if data["total"] != float64(53) || len(data["agents"].([]any)) != 3 || calls.Load() != 2 {
		t.Fatal("page auto-followed or truncated", data, calls.Load())
	}
	for _, cmd := range contextcli.Capabilities("fixture").Commands {
		if cmd.Name == "agents list" || cmd.Name == "agents show" {
			if len(cmd.Modes) != 2 || cmd.Status != contextcli.CommandAvailable {
				t.Fatal("user capability missing", cmd)
			}
		}
	}
}
