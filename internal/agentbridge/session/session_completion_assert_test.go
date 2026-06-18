package session

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func assertTextDeltaSeen(t *testing.T, events []agentbridge.Event, want string) {
	t.Helper()
	for _, ev := range events {
		if ev.Kind == agentbridge.EventTextDelta && ev.Text == want {
			return
		}
	}
	t.Fatalf("expected to see TextDelta %q in event stream, got %+v", want, events)
}

func assertProgressSeen(t *testing.T, events []agentbridge.Event, want string) {
	t.Helper()
	for _, ev := range events {
		if ev.Kind == agentbridge.EventProgress && ev.Text == want {
			return
		}
	}
	t.Fatalf("missing progress event %q in %+v", want, events)
}
