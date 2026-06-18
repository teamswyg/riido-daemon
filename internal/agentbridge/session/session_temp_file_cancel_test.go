package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionRemovesTempFilesAfterCancellation(t *testing.T) {
	tempFilePath := newSessionTempFilePath(t)
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	scenario := startToolGateScenario(t, "task-tempfile-cancel", adapter, func(cfg *Config) {
		cfg.TempFiles = []string{tempFilePath}
	})

	scenario.session.Cancel(nil)
	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("result: %+v", res)
	}
	assertSessionTempFileRemoved(t, tempFilePath, "cancellation")
	_ = drainEvents(t, scenario.session, time.Second)
}
