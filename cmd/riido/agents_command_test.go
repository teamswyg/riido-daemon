package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestAgentsCommandsForwardLiveDirectoryReads(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"daemon-1","data":{"agents":[]},"sources":[],"warnings":[],"page":null,"error":null}`
	if err := runAgents(h.io, []string{"list", "--cursor", "1", "--json"}); err != nil {
		t.Fatalf("agents list: %v", err)
	}
	req := h.daemon.lastRequest(t)
	args, _ := req["args"].(map[string]any)
	if req["command"] != "agents list" || args["cursor"] != "1" {
		t.Fatalf("list request: %v", req)
	}
	if err := runAgents(h.io, []string{"show", "agent-b", "--json"}); err != nil {
		t.Fatalf("agents show: %v", err)
	}
	req = h.daemon.lastRequest(t)
	args, _ = req["args"].(map[string]any)
	if req["command"] != "agents show" || args["agent"] != "agent-b" {
		t.Fatalf("show request: %v", req)
	}
	if strings.Contains(h.stdout.String()+h.stderr.String(), testRunCapability) {
		t.Fatal("run capability printed")
	}
}

func TestAgentsCommandUsageErrors(t *testing.T) {
	for _, argv := range [][]string{{}, {"show", "--json"}, {"delete", "--json"}} {
		h := newContextHarness(t, agentVars())
		if code := exitCodeOf(t, runAgents(h.io, argv)); code != 2 {
			t.Fatalf("%v: exit %d", argv, code)
		}
	}
	h := newContextHarness(t, agentVars())
	if err := runAgents(h.io, []string{"--help"}); err != nil || !strings.Contains(h.stderr.String(), "riido agents show") {
		t.Fatalf("help: %v %s", err, h.stderr.String())
	}
}
