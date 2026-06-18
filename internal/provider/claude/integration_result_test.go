package claude

import (
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
		t.Fatalf("claude integration did not complete: %+v", res)
	}
}
