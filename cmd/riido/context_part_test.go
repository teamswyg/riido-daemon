package main

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/contextcli"
	"github.com/teamswyg/riido-daemon/internal/standalonecli"
)

func TestContextPartUsesOnlyAgentCapturedPage(t *testing.T) {
	h := newContextHarness(t, agentVars())
	h.daemon.replies[contextcli.IPCMethodContextRead] = `{"schema_version":"1","request_id":"fixture","data":{"content":"captured"}}`
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("agent content read opened a user session")
		return nil, nil
	}
	if err := runContext(h.io, []string{"part", "work-one", "--cursor", "captured-page", "--json"}); err != nil {
		t.Fatal(err)
	}
	if len(h.daemon.requests) != 1 {
		t.Fatal("expected one broker request")
	}
	req := h.daemon.lastRequest(t)
	args := req["args"].(map[string]any)
	if req["command"] != "context part" || args["work"] != "work-one" || args["cursor"] != "captured-page" {
		t.Fatalf("wrong captured part request: %+v", req)
	}
}

func TestContextPartRequiresSourceAndCursor(t *testing.T) {
	for _, args := range [][]string{
		{"part"},
		{"part", "work-one"},
		{"part", "work-one", "--cursor", ""},
		{"part", "work-one", "--cursor", " "},
		{"part", "RIID-5900", "--cursor", "page"},
	} {
		h := newContextHarness(t, agentVars())
		if code := exitCodeOf(t, runContext(h.io, args)); code != 2 {
			t.Fatalf("%v: exit %d, want 2", args, code)
		}
		if len(h.daemon.requests) != 0 {
			t.Fatal("invalid page request reached broker")
		}
	}
}

func TestUserContextPartNeverOpensCredentialsOrDaemon(t *testing.T) {
	h := newContextHarness(t, nil)
	h.io.userSession = func(contextcli.Args) (*standalonecli.Service, error) {
		t.Fatal("snapshot cursor opened a user session")
		return nil, nil
	}
	if code := exitCodeOf(t, runContext(h.io, []string{"part", "work-one", "--cursor", "page"})); code != 8 {
		t.Fatalf("user snapshot cursor exit = %d, want 8", code)
	}
	if len(h.daemon.requests) != 0 {
		t.Fatal("user snapshot cursor reached daemon")
	}
}
