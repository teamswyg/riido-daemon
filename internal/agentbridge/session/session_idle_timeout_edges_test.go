package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionSemanticIdleTimeoutAfterProgress(t *testing.T) {
	adapter := &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) != "PROGRESS" {
				return nil, nil, nil
			}
			return []agentbridge.Event{{Kind: agentbridge.EventProgress, Text: "working"}}, nil, nil
		},
	}
	scenario := startToolGateScenario(t, "task-semantic-idle", adapter, func(cfg *Config) {
		cfg.SemanticIdle = 40 * time.Millisecond
		cfg.ProcessKillTimeout = 10 * time.Millisecond
	})

	scenario.running.EmitStdout([]byte("PROGRESS"))
	res := waitResult(t, scenario.session, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("status=%s", res.Status)
	}
	if res.Error != "semantic idle timeout" {
		t.Fatalf("error=%q", res.Error)
	}
	expectTimeoutKill(t, scenario)
	events := drainEvents(t, scenario.session, time.Second)
	if !hasProgressText(events, "working") {
		t.Fatalf("missing progress event: %+v", events)
	}
}

func hasProgressText(events []agentbridge.Event, text string) bool {
	for _, ev := range events {
		if ev.Kind == agentbridge.EventProgress && ev.Text == text {
			return true
		}
	}
	return false
}
