package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionRemovesTempFilesAfterTimeout(t *testing.T) {
	tempFilePath := newSessionTempFilePath(t)
	scenario := startToolGateScenario(t, "task-tempfile-timeout", noEventRecordingAdapter(), func(cfg *Config) {
		cfg.HardTimeout = 20 * time.Millisecond
		cfg.TempFiles = []string{tempFilePath}
	})

	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("result: %+v", res)
	}
	assertSessionTempFileRemoved(t, tempFilePath, "timeout")
	_ = drainEvents(t, scenario.session, time.Second)
}
