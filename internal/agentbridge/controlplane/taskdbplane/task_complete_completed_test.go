package taskdbplane

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCompleteTaskSynthesizesRunStartedWhenProviderOmitsLifecycle(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateValidating {
		t.Fatalf("state = %s, want Validating", got)
	}
	assertTransition(t, db, ir.EventRunStarted)
	assertTransition(t, db, ir.EventRunReportedDone)
}

func TestCompleteTaskCompletedReplayDoesNotAppendDuplicateRunDone(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatalf("CompleteTask first: %v", err)
	}
	before := loadTaskDB(t, path)
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted}); err != nil {
		t.Fatalf("CompleteTask replay: %v", err)
	}
	after := loadTaskDB(t, path)
	if len(after.Transitions) != len(before.Transitions) || len(after.CommandReceipts) != len(before.CommandReceipts) {
		t.Fatalf("replay appended mutation: before=%d/%d after=%d/%d", len(before.Transitions), len(before.CommandReceipts), len(after.Transitions), len(after.CommandReceipts))
	}
}
