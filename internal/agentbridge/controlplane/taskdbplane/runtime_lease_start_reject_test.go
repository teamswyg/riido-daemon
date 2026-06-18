package taskdbplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestStartTaskRejectsExpiredRuntimeLease(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	now := time.Date(2026, 5, 25, 4, 0, 0, 0, time.UTC)
	plane.now = func() time.Time { return now }
	reportCtx := writeRuntimeLease(t, plane, RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:task-1:1",
		TaskID:                "task-1",
		RuntimeID:             "runtime-1",
		CapabilityFingerprint: "runtime-1-fp",
		ClaimedAt:             now.Add(-time.Hour),
		LeaseUntil:            now.Add(-time.Minute),
		FencingToken:          1,
	})

	if err := plane.StartTask(reportCtx, "task-1"); err == nil {
		t.Fatal("StartTask should reject expired runtime lease")
	}
	assertTaskState(t, path, "task-1", task.StateClaimed)
}

func TestStartTaskRejectsMismatchedFencingToken(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	reportCtx := writeActiveRuntimeLease(t, plane, "task-1")
	report, ok := controlplane.TaskReportContextFromContext(reportCtx)
	if !ok {
		t.Fatal("missing task report context")
	}
	report.RuntimeFencingToken = 999
	staleCtx := controlplane.ContextWithTaskReport(context.Background(), report)

	if err := plane.StartTask(staleCtx, "task-1"); err == nil {
		t.Fatal("StartTask should reject mismatched fencing token")
	}
	assertTaskState(t, path, "task-1", task.StateClaimed)
}
