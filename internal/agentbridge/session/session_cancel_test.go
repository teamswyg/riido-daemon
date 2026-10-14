package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionCancellation(t *testing.T) {
	scenario := startToolGateScenario(t, "task-3", noEventRecordingAdapter(), nil)

	scenario.session.Cancel(nil)
	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("status: %s", res.Status)
	}
}
