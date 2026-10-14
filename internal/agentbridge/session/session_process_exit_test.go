package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionProcessExitNonZero(t *testing.T) {
	scenario := startToolGateScenario(t, "task-4", noEventRecordingAdapter(), nil)
	go scenario.running.EmitExit(137, nil)

	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultFailed {
		t.Fatalf("status: %s", res.Status)
	}
}
