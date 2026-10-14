package codex

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestTranslateTurnStarted(t *testing.T) {
	raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"turn_started","params":{}}`)
	evs := tx(t, raw)
	if len(evs) != 1 || evs[0].Kind != agentbridge.EventLifecycle || evs[0].Phase != agentbridge.StateRunning {
		t.Fatalf("turn_started: %+v", evs)
	}
}

func TestTranslateNewLifecycleNotificationsAreNonTerminalLogs(t *testing.T) {
	methods := []string{
		"item/started", "item/updated", "item/completed",
		"hook/started", "hook/completed",
		"mcpServer/startupStatus/updated",
		"remoteControl/status/changed",
		"account/rateLimits/updated",
	}
	for _, m := range methods {
		raw := rawFromJSON(t, `{"jsonrpc":"2.0","method":"`+m+`","params":{}}`)
		evs := tx(t, raw)
		if len(evs) != 1 || evs[0].Kind != agentbridge.EventLog {
			t.Fatalf("%s: want single EventLog, got %+v", m, evs)
		}
		if strings.Contains(evs[0].Text, "unknown") {
			t.Fatalf("%s: still labeled unknown: %q", m, evs[0].Text)
		}
	}
}
