package codex

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// Newer codex app-server builds end a turn with thread/status/changed instead of
// turn/completed. The driver must complete the run on a terminal status, but
// only after a turn actually started (so an initial idle can't end it early),
// fail on an error status, and never hard-fail on an unknown status.
func TestThreadStatusEvents(t *testing.T) {
	// terminal status before any turn → benign log, no completion.
	d := &protocolDriver{}
	ev := d.threadStatusEvents(map[string]any{"status": "idle"})
	if len(ev) != 1 || ev[0].Kind != agentbridge.EventLog {
		t.Fatalf("idle before a turn must not complete the run: %+v", ev)
	}

	// active status → running lifecycle + marks the turn active.
	ev = d.threadStatusEvents(map[string]any{"status": "running"})
	if len(ev) != 1 || ev[0].Kind != agentbridge.EventLifecycle {
		t.Fatalf("active status should map to a running lifecycle: %+v", ev)
	}
	if !d.turnStarted {
		t.Fatal("active status should mark the turn started")
	}

	// terminal status after a turn → completed result.
	ev = d.threadStatusEvents(map[string]any{"status": "completed"})
	if len(ev) != 1 || ev[0].Kind != agentbridge.EventResult ||
		ev[0].Result.Status != agentbridge.ResultCompleted {
		t.Fatalf("terminal status after a turn should complete: %+v", ev)
	}

	// error status → failed result (regardless of turn state).
	d2 := &protocolDriver{turnStarted: true}
	ev = d2.threadStatusEvents(map[string]any{"status": "failed"})
	last := ev[len(ev)-1]
	if last.Kind != agentbridge.EventResult || last.Result.Status != agentbridge.ResultFailed {
		t.Fatalf("error status should fail the run: %+v", ev)
	}

	// unknown status → log only, never a failure.
	d3 := &protocolDriver{turnStarted: true}
	ev = d3.threadStatusEvents(map[string]any{"status": "something-new"})
	if len(ev) != 1 || ev[0].Kind != agentbridge.EventLog {
		t.Fatalf("unknown status must not fail the run: %+v", ev)
	}
}
