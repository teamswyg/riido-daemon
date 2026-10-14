package session

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionRemovesTempFilesAfterProcessExit(t *testing.T) {
	tempFilePath := newSessionTempFilePath(t)
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) != "DONE" {
				return nil, nil, nil
			}
			return []agentbridge.Event{{Kind: agentbridge.EventResult, Result: completedResult()}}, nil, nil
		},
	}
	scenario := startToolGateScenario(t, "task-tempfile-exit", adapter, func(cfg *Config) {
		cfg.TempFiles = []string{tempFilePath, tempFilePath}
	})
	go func() {
		scenario.running.EmitStdout([]byte("DONE"))
		scenario.running.EmitExit(0, nil)
	}()

	res := waitResult(t, scenario.session, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	assertSessionTempFileRemoved(t, tempFilePath, "process exit")
	_ = drainEvents(t, scenario.session, time.Second)
}

func completedResult() agentbridge.Result {
	return agentbridge.Result{Status: agentbridge.ResultCompleted}
}

func assertSessionTempFileRemoved(t *testing.T, path, reason string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("temp file should be removed after %s, stat err=%v", reason, err)
	}
}
