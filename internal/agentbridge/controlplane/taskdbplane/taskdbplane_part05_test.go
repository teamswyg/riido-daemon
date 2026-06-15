package taskdbplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func TestNewLoadsPersistedRuntimeRegistryForClaimGating(t *testing.T) {
	path := writeTaskDB(t, taskdb.TaskDB{
		SchemaVersion:       taskdb.TaskDBSchemaVersion,
		RecommendedProvider: "codex",
		ProviderCandidates: []taskdb.ProviderCandidate{
			{ID: "codex", Available: true},
			{ID: "claude", Available: true},
		},
		Tasks: []taskdb.TaskRecord{
			{
				ID:                  "codex-task",
				ProjectID:           "project-1",
				State:               task.StateQueued,
				Title:               "codex task",
				RecommendedProvider: "codex",
				UpdatedAt:           "2026-05-25T00:00:00Z",
			},
			{
				ID:                  "claude-task",
				ProjectID:           "project-1",
				State:               task.StateQueued,
				Title:               "claude task",
				RecommendedProvider: "claude",
				UpdatedAt:           "2026-05-25T00:00:01Z",
			},
		},
	})
	plane := newTestPlane(t, path)
	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		RuntimeID: "runtime-claude",
		Provider:  "multi",
		Capabilities: map[string]bool{
			"provider.codex.available":  false,
			"provider.claude.available": true,
		},
	}); err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}

	reloaded, err := New(path)
	if err != nil {
		t.Fatalf("New reload: %v", err)
	}
	reloaded.now = plane.now
	req, err := reloaded.ClaimTask(context.Background(), "runtime-claude")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != "claude-task" {
		t.Fatalf("reloaded runtime registry should gate claim, got %+v", req)
	}
}

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

func newTestPlane(t *testing.T, path string) *Plane {
	t.Helper()
	plane, err := New(path)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	now := time.Date(2026, 5, 25, 1, 2, 3, 0, time.UTC)
	plane.now = func() time.Time {
		now = now.Add(time.Second)
		return now
	}
	return plane
}
