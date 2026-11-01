package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

const userAgentFixture = `{"agent_id":"agent-one","name":"리뷰 Agent","description":"공개 설명","role_status":"described","work_status":"idle","runtime":{"kind":"codex","connection":"offline"},"callable":false,"unavailable_reasons":["runtime_offline"]}`

func TestUserAgentsCLIHTTPListAndShow(t *testing.T) {
	for _, command := range []string{"list", "show"} {
		t.Run(command, func(t *testing.T) {
			body := `{"agents":[` + userAgentFixture + `],"total":1,"cursor":"0","fetched_at":"2026-11-01T01:00:00Z","policy_checked":false}`
			args := []string{command}
			if command == "show" {
				body = `{"agent":` + userAgentFixture + `,"fetched_at":"2026-11-01T01:00:00Z","policy_checked":false}`
				args = append(args, "agent-one")
			}
			s, calls := userAgentsCLIService(t, body, 200)
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) { return s, nil }
			args = append(args, "--work", "task-one", "--workspace", "workspace-one", "--json")
			if err := runAgents(h.io, args); err != nil {
				t.Fatal(err)
			}
			data := decodeEnvelope(t, h.stdout.Bytes()).Data.(map[string]any)
			directory := data["directory"].(map[string]any)
			if data["mode"] != "user" || data["live"] != true || directory["policy_checked"] != false || calls.Load() != 2 {
				t.Fatal("lost user authority or fresh two-hop request")
			}
			for _, secret := range []string{"fixture-access", "fixture-refresh", "riido_cli_agent_v1_"} {
				if strings.Contains(h.stdout.String()+h.stderr.String(), secret) {
					t.Fatal("credential exposed")
				}
			}
		})
	}
}
