package taskdbplane

import (
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestReporterTransitionsRunLifecycleAndTerminalDone(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")

	if err := plane.StartTask(reportCtx, "task-1"); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.ReportEvent(reportCtx, "task-1", agentbridge.Event{Kind: agentbridge.EventLifecycle, Phase: agentbridge.StateRunning}); err != nil {
		t.Fatalf("ReportEvent: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, "task-1", agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "done"}); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}

	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateValidating {
		t.Fatalf("state = %s, want Validating", got)
	}
	assertTransition(t, db, ir.EventWorkdirPreparing)
	assertTransition(t, db, ir.EventRunStarted)
	assertTransition(t, db, ir.EventRunReportedDone)
}
