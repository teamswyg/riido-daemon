package taskdbplane

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCompleteTaskFailedFromPreparing(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultFailed, Error: "boom"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateFailed {
		t.Fatalf("state = %s, want Failed", got)
	}
	assertTransition(t, db, ir.EventTaskFailed)
}

func TestCompleteTaskBlockedFromPreparing(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultBlocked, Error: "missing required surface"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateBlocked {
		t.Fatalf("state = %s, want Blocked", got)
	}
	assertTransition(t, db, ir.EventBlockerRaised)
}
