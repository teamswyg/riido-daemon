package session

import (
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionProcessExitZeroWithoutProviderResultAborts(t *testing.T) {
	scenario := startToolGateScenario(t, "task-exit-zero-no-result", noEventRecordingAdapter(), nil)
	go scenario.running.EmitExit(0, nil)

	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultAborted {
		t.Fatalf("status=%s", res.Status)
	}
	if res.Error != "process exited without provider result" {
		t.Fatalf("error=%q", res.Error)
	}
}

func TestSessionProcessExitNonZeroPreservesError(t *testing.T) {
	scenario := startToolGateScenario(t, "task-exit-nonzero-error", noEventRecordingAdapter(), nil)
	go scenario.running.EmitExit(2, errors.New("provider crashed"))

	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultFailed {
		t.Fatalf("status=%s", res.Status)
	}
	if res.Error != "provider crashed" {
		t.Fatalf("error=%q", res.Error)
	}
}
