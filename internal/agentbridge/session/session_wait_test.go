package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func drainEvents(t *testing.T, sess *Session, deadline time.Duration) []agentbridge.Event {
	t.Helper()
	var out []agentbridge.Event
	timer := time.NewTimer(deadline)
	defer timer.Stop()
	for {
		select {
		case ev, ok := <-sess.Events():
			if !ok {
				return out
			}
			out = append(out, ev)
		case <-timer.C:
			t.Fatal("drainEvents deadline exceeded")
			return out
		}
	}
}

func waitResult(t *testing.T, sess *Session, deadline time.Duration) agentbridge.Result {
	t.Helper()
	select {
	case res := <-sess.Result():
		return res
	case <-time.After(deadline):
		t.Fatal("waitResult deadline exceeded")
		return agentbridge.Result{}
	}
}
