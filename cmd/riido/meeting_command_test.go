package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
)

func TestMeetingCommandHelpIsRegistered(t *testing.T) {
	if err := run([]string{"meeting", "--help"}); err != nil {
		t.Fatalf("meeting help is not registered: %v", err)
	}
}

func TestMeetingCommandsForwardScopedReads(t *testing.T) {
	for _, tc := range []struct {
		argv           []string
		command, key   string
		target, budget string
	}{
		{[]string{"list", "--json"}, "meeting list", "work", "", ""},
		{[]string{"list", "--work", "riid-5900", "--json"}, "meeting list", "work", "RIID-5900", ""},
		{[]string{"get", "m1", "--budget", "1200", "--json"}, "meeting get", "meeting", "m1", "1200"},
	} {
		t.Run(strings.Join(tc.argv, " "), func(t *testing.T) {
			h := newContextHarness(t, agentVars())
			h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"daemon-1","data":{},"sources":[],"warnings":[],"page":null,"error":null}`
			if err := runMeeting(h.io, tc.argv); err != nil {
				t.Fatal(err)
			}
			req := h.daemon.lastRequest(t)
			args, _ := req["args"].(map[string]any)
			if req["command"] != tc.command || args[tc.key] != tc.target || (tc.budget != "" && args["budget_chars"] != tc.budget) {
				t.Fatalf("meeting request: %v", req)
			}
			if strings.Contains(h.stdout.String()+h.stderr.String(), testRunCapability) {
				t.Fatal("capability was printed")
			}
		})
	}
}

func TestMeetingCommandRejectsInvalidOrUnavailableUse(t *testing.T) {
	for _, argv := range [][]string{{}, {"get", "--json"}, {"get", " ", "--json"}, {"get", "m1", "--budget", "0", "--json"}, {"list", "extra", "--json"}, {"list", "--work", "../private", "--json"}, {"delete", "m1", "--json"}} {
		h := newContextHarness(t, agentVars())
		if got := exitCodeOf(t, runMeeting(h.io, argv)); got != 2 {
			t.Fatalf("%v: exit %d", argv, got)
		}
	}
	h := newContextHarness(t, nil)
	if got := exitCodeOf(t, runMeeting(h.io, []string{"get", "m1", "--json"})); got != 2 {
		t.Fatalf("user meeting get requires explicit team before authentication, exit %d", got)
	}
	if got := exitCodeOf(t, runMeeting(h.io, []string{"list", "--json"})); got != 8 || len(h.daemon.requests) != 0 {
		t.Fatalf("user meeting list remains unavailable without daemon fallback, exit %d", got)
	}
}
