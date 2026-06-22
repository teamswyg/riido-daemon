package claude

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/session"
)

func drainClaudeIntegrationEvents(sess *session.Session) {
	go func() {
		for range sess.Events() {
		}
	}()
}

func requireClaudeIntegrationCompleted(t *testing.T, res agentbridge.Result) {
	t.Helper()

	if res.Status != agentbridge.ResultCompleted {
		if claudeAuthMissing(res) {
			t.Skip("claude authentication missing; run claude /login or configure API credentials")
		}
		t.Fatalf("claude integration did not complete: %+v", res)
	}
}

func claudeAuthMissing(res agentbridge.Result) bool {
	text := strings.ToLower(res.Error + " " + res.Output)
	return strings.Contains(text, "not logged in") ||
		strings.Contains(text, "authentication")
}
