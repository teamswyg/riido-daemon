package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserRelationsArgumentsFailBeforeStorage(t *testing.T) {
	base := []string{"relations", "work-one", "--team-id", "team-one"}
	for _, tc := range []struct {
		name string
		args []string
		exit int
	}{
		{"missing-work", []string{"relations", "--team-id", "team-one"}, 2},
		{"missing-team", []string{"relations", "work-one"}, 2},
		{"key", []string{"relations", "RIID-5900", "--team-id", "team-one"}, 8},
		{"team-key", []string{"relations", "work-one", "--team-id", "RIID-5900"}, 8},
		{"empty-team", []string{"relations", "work-one", "--team-id="}, 2},
		{"empty-cursor", append(slices.Clone(base), "--cursor="), 8},
		{"empty-workspace", append(slices.Clone(base), "--workspace="), 8},
		{"page", append(slices.Clone(base), "--page", "1"), 8},
		{"limit", append(slices.Clone(base), "--limit", "20"), 8},
		{"secret", append(slices.Clone(base), "--token", "private-token"), 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Fatal("invalid arguments opened storage")
				return nil, nil
			}
			if got := exitCodeOf(t, runWork(h.io, append(tc.args, "--json"))); got != tc.exit {
				t.Fatalf("exit=%d want=%d", got, tc.exit)
			}
			if len(h.daemon.requests) != 0 {
				t.Fatal("invalid user command used daemon")
			}
		})
	}
}

func TestUserRelationsHelpCapabilitiesAndAgentSnapshot(t *testing.T) {
	h := newContextHarness(t, nil)
	if err := runWork(h.io, []string{"relations", "--help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(h.stderr.String(), "relations <ID> --team-id") {
		t.Fatal("user relation help missing")
	}
	commands := contextcli.Capabilities("test").Commands
	i := slices.IndexFunc(commands, func(c contextcli.CommandCapability) bool { return c.Name == "work relations" })
	if i < 0 || commands[i].Status != contextcli.CommandAvailable || !slices.Contains(commands[i].Modes, contextcli.ModeUser) || !slices.Contains(commands[i].Modes, contextcli.ModeAgent) {
		t.Fatal("relation capability modes changed")
	}
	h = newContextHarness(t, agentVars())
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("agent opened user storage")
		return nil, nil
	}
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"fixture","data":{"relations":[]},"sources":[],"warnings":[{"code":"ASSIGNMENT_SNAPSHOT","message":"fixture"}],"page":null,"error":null}`
	if err := runWork(h.io, []string{"relations", "RIID-5900", "--json"}); err != nil {
		t.Fatal(err)
	}
	request := h.daemon.lastRequest(t)
	if request["command"] != "work relations" || request["args"].(map[string]any)["work"] != "RIID-5900" {
		t.Fatal("agent snapshot request changed")
	}
}
