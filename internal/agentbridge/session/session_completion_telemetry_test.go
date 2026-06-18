package session

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionExtractsRiidoTelemetryAboveAdapters(t *testing.T) {
	started := startRecordingSession(t, "task-telemetry", telemetryCompletionAdapter(), nil)
	go func() {
		started.running.EmitStdout([]byte("<riido_log>프로젝트 go"))
		started.running.EmitStdout([]byte(".mod 작성중<end>"))
		started.running.EmitStdout([]byte("DONE"))
		started.running.EmitExit(0, nil)
	}()
	res := waitResult(t, started.sess, 2*time.Second)
	if res.Status != agentbridge.ResultCompleted {
		t.Fatalf("result: %+v", res)
	}
	assertProgressSeen(t, drainEvents(t, started.sess, time.Second), "프로젝트 go.mod 작성중")
}

func telemetryCompletionAdapter() *recordingAdapter {
	return &recordingAdapter{
		name:   "fake",
		parser: &recordingParser{},
		translateFn: func(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
			if string(raw.Bytes) == "DONE" {
				return []agentbridge.Event{completedResultEvent("")}, nil, nil
			}
			return []agentbridge.Event{
				{Kind: agentbridge.EventTextDelta, Text: string(raw.Bytes)},
			}, nil, nil
		},
	}
}
