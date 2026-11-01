package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestContextChangesForwardsOnlyComparisonOptions(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"test","data":{"changes":[]},"error":null}`
	if err := runContext(h.io, []string{"changes", "--since", "previous", "--budget", "2048", "--json"}); err != nil {
		t.Fatal(err)
	}
	req := h.daemon.lastRequest(t)
	args, _ := req["args"].(map[string]any)
	if req["command"] != "context changes" || args["since"] != "previous" || args["budget_chars"] != "2048" || len(args) != 2 {
		t.Fatal("wrong comparison IPC")
	}
}

func TestContextChangesInvalidSelectorsNeverReachIPC(t *testing.T) {
	for _, args := range [][]string{{}, {"--since", ""}, {"--since", strings.Repeat("a", 257)}, {"--since", "x\ny"}, {"--since", "\nx"}, {"--since", "x", "--cursor="}, {"--since", "x", "--cursor", "c"}, {"--since", "x", "--work", "other"}, {"--since", "x", "--budget", "0"}, {"--since", "x", "--budget="}} {
		h := newContextHarness(t, agentVars())
		raw := append([]string{"changes", "--json"}, args...)
		if exitCodeOf(t, runContext(h.io, raw)) != 2 || len(h.daemon.requests) != 0 {
			t.Fatal("invalid changes input reached IPC")
		}
	}
}
