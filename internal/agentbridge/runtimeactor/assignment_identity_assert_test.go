package runtimeactor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func assertRunningAssignmentIDs(t *testing.T, a *Actor, want ...string) {
	t.Helper()
	status, err := a.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.RunningSessions != len(want) {
		t.Fatalf("running sessions = %d, want %d", status.RunningSessions, len(want))
	}
	seen := map[string]bool{}
	for _, task := range status.RunningTasks {
		seen[task.TaskID] = true
	}
	for _, id := range want {
		if !seen[id] {
			t.Fatalf("running task ids = %+v, missing %q", status.RunningTasks, id)
		}
	}
	if seen[assignmentIdentityLogicalTaskID] {
		t.Fatalf("running task ids = %+v, want assignment ids only", status.RunningTasks)
	}
}

func assertHeartbeatAssignmentIDs(t *testing.T, a *Actor, want ...string) {
	t.Helper()
	hb, err := a.HeartbeatPayload(context.Background())
	if err != nil {
		t.Fatalf("heartbeat: %v", err)
	}
	got := hb.RunningTaskIDs
	if len(got) != len(want) {
		t.Fatalf("heartbeat running ids = %v, want %v", got, want)
	}
	for i, id := range want {
		if got[i] != id {
			t.Fatalf("heartbeat running ids = %v, want %v", got, want)
		}
	}
}

func assertLogicalTaskCancelRejected(t *testing.T, a *Actor) {
	t.Helper()
	err := a.Cancel(context.Background(), assignmentIdentityLogicalTaskID, "logical id is not execution id")
	if !errors.Is(err, ErrUnknownTask) {
		t.Fatalf("cancel by logical task id = %v, want ErrUnknownTask", err)
	}
}

func assertAssignmentKilled(t *testing.T, r *process.FakeRunning) {
	t.Helper()
	select {
	case <-r.KillRecv():
	case <-time.After(2 * time.Second):
		t.Fatal("second assignment process was not killed")
	}
}
