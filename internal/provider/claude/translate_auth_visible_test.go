package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateLogClaudeAuthGuidesLogin(t *testing.T) {
	raw := mustParseRaw(t, `{"type":"log","message":"Not logged in · Please run /login"}`)
	events := translate(t, raw)
	if len(events) != 1 || events[0].Kind != agentbridge.EventLog {
		t.Fatalf("events: %+v", events)
	}
	if strings.Contains(events[0].Text, "Please run /login") ||
		!strings.Contains(events[0].Text, "claude auth login") {
		t.Fatalf("message: %q", events[0].Text)
	}
}
