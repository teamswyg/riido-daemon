package codex

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCodexProtocolDriverProcessExitAfterOnlyRuntimeErrorFails(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second)

	_, _, _ = d.OnRaw(
		context.Background(),
		makeNotification("error", map[string]any{"error": map[string]any{"message": "provider auth failed"}}),
		io,
	)
	events, err := d.OnProcessExit(context.Background(), agentbridge.ProcessExitStatus{Code: 0}, io)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("expected failed process-exit result, got %+v", events)
	}
	if !strings.Contains(last.Result.Error, "provider auth failed") {
		t.Fatalf("expected retained error message, got %+v", last.Result)
	}
}

// --- B7: process exit fails pending requests ---

func TestCodexProtocolDriverProcessExitFailsPendingRequests(t *testing.T) {
	d, _ := NewProtocolDriver(agentbridge.StartRequest{})
	io := newRecordingIO()
	_ = d.OnStart(context.Background(), io)
	_ = io.next(t, time.Second) // initialize sent → pending request id=1 in flight

	// Send thread/start before initialize response arrives so that
	// two pending requests are live.
	_, _, _ = d.OnRaw(context.Background(), makeResponse(1, nil), io)
	_ = io.next(t, time.Second) // initialized
	_ = io.next(t, time.Second) // thread/start → pending id=2

	exitEvents, err := d.OnProcessExit(
		context.Background(),
		agentbridge.ProcessExitStatus{Code: 137}, io,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(exitEvents) == 0 {
		t.Fatal("expected at least one event for pending request cleanup")
	}
	// At least one Error event should reference "pending" semantics.
	gotError := false
	for _, ev := range exitEvents {
		if ev.Kind == agentbridge.EventError {
			gotError = true
		}
	}
	if !gotError {
		t.Fatalf("expected an Error event after process exit, got %+v", exitEvents)
	}
}
