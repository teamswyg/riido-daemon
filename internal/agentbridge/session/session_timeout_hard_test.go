package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionHardTimeout(t *testing.T) {
	scenario := startToolGateScenario(t, "task-2", noEventRecordingAdapter(), func(cfg *Config) {
		cfg.HardTimeout = 100 * time.Millisecond
	})

	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("status: %s", res.Status)
	}
	expectTimeoutKill(t, scenario)
}

func expectTimeoutKill(t *testing.T, scenario toolGateScenario) {
	t.Helper()
	select {
	case <-scenario.running.KillRecv():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected kill on timeout")
	}
}
