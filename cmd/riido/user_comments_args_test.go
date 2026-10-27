package main

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestUserCommentArgumentsFailBeforeProtectedStorage(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		exit int
	}{
		{"missing-team", []string{"comments", "work-one"}, 2},
		{"work-key", []string{"comments", "RIID-5900", "--team-id", "team-one"}, 8},
		{"page-zero", []string{"comments", "work-one", "--team-id", "team-one", "--page", "0"}, 2},
		{"page-empty", []string{"comments", "work-one", "--team-id", "team-one", "--page="}, 2},
		{"page-overflow", []string{"comments", "work-one", "--team-id", "team-one", "--page", "2147483648"}, 2},
		{"offset-overflow", []string{"comments", "work-one", "--team-id", "team-one", "--page", "214748365", "--limit", "20"}, 2},
		{"limit-zero", []string{"comments", "work-one", "--team-id", "team-one", "--limit", "0"}, 2},
		{"limit-empty", []string{"comments", "work-one", "--team-id", "team-one", "--limit="}, 2},
		{"limit-too-large", []string{"comments", "work-one", "--team-id", "team-one", "--limit", "101"}, 2},
		{"cursor", []string{"comments", "work-one", "--team-id", "team-one", "--cursor", "opaque"}, 8},
		{"empty-cursor", []string{"comments", "work-one", "--team-id", "team-one", "--cursor="}, 8},
		{"reply-missing-parent", []string{"replies", "work-one", "--team-id", "team-one"}, 2},
		{"reply-limit", []string{"replies", "work-one", "--team-id", "team-one", "--comment-id", "comment-one", "--limit", "20"}, 8},
		{"reply-page", []string{"replies", "work-one", "--team-id", "team-one", "--comment-id", "comment-one", "--page", "1"}, 8},
		{"secret", []string{"replies", "work-one", "--token", "fake-secret"}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newContextHarness(t, nil)
			h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
				t.Fatal("invalid arguments reached protected storage")
				return nil, nil
			}
			if got := exitCodeOf(t, runWork(h.io, append(tc.args, "--json"))); got != tc.exit {
				t.Fatalf("exit=%d want=%d", got, tc.exit)
			}
			if len(h.daemon.requests) != 0 {
				t.Fatal("invalid user command reached daemon")
			}
		})
	}
}

func TestAgentCommentCursorIsUnchangedAndRepliesRemainUnsupported(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("agent opened user storage")
		return nil, nil
	}
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"fixture","data":{"comments":[]},"sources":[],"warnings":[],"page":{"next_cursor":"next-opaque"},"error":null}`
	if err := runWork(h.io, []string{"comments", "RIID-5900", "--cursor", "previous-opaque", "--json"}); err != nil {
		t.Fatal(err)
	}
	request := h.daemon.lastRequest(t)
	args := request["args"].(map[string]any)
	if request["command"] != "work comments" || args["cursor"] != "previous-opaque" || args["limit"] != nil {
		t.Fatalf("agent paging changed: %v", args)
	}
	for _, command := range [][]string{{"comments", "RIID-5900", "--limit", "20"}, {"replies", "work-one", "--team-id", "team-one", "--comment-id", "comment-one"}} {
		if got := exitCodeOf(t, runWork(h.io, command)); got != 8 {
			t.Fatalf("agent unsupported exit=%d", got)
		}
	}
	if len(h.daemon.requests) != 1 {
		t.Fatal("unsupported command reached daemon")
	}
}

func TestUserCommentHelpAndCapabilitiesDescribeModeBoundaries(t *testing.T) {
	h := newContextHarness(t, nil)
	if err := runWork(h.io, []string{"comments", "--help"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"--page", "--limit", "--comment-id", "agent replies unsupported"} {
		if !strings.Contains(h.stderr.String(), want) {
			t.Fatalf("missing help: %s", want)
		}
	}
	for _, name := range []string{"work comments", "work replies"} {
		found := false
		for _, capability := range contextcli.Capabilities("fixture").Commands {
			if capability.Name != name {
				continue
			}
			for _, mode := range capability.Modes {
				if mode == contextcli.ModeUser {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("missing user capability: %s", name)
		}
	}
}
