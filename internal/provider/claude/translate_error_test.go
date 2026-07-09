package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateErrorEvent(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"error","message":"provider exploded"}`)
	events := translate(t, raw)

	if len(events) != 1 {
		t.Fatalf("events: %+v", events)
	}
	if events[0].Kind != agentbridge.EventError {
		t.Fatalf("kind: %s", events[0].Kind)
	}
	if events[0].Err != "provider exploded" {
		t.Fatalf("err: %q", events[0].Err)
	}
}

func TestTranslateErrorEventClaude401GuidesReauthentication(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"error","message":"Failed to authenticate, API Error: 401 Invalid authentication credentials"}`)
	events := translate(t, raw)

	if len(events) != 1 || events[0].Kind != agentbridge.EventError {
		t.Fatalf("events: %+v", events)
	}
	if !strings.Contains(events[0].Err, "Claude Code 인증") ||
		!strings.Contains(events[0].Err, "claude auth login") {
		t.Fatalf("err: %q", events[0].Err)
	}
}
