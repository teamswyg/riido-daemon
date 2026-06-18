package taskdbplane

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestCompleteTaskReleasesRuntimeLease(t *testing.T) {
	path := writeQueuedCodexTaskDB(t)
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)
	req := claimRuntimeTask(t, plane, "runtime-1")
	reportCtx := contextWithTaskRequest(t, req)
	if err := plane.StartTask(reportCtx, codexTaskID); err != nil {
		t.Fatalf("StartTask: %v", err)
	}
	if err := plane.CompleteTask(reportCtx, codexTaskID, failedRuntimeResult()); err != nil {
		t.Fatalf("CompleteTask: %v", err)
	}
	assertRuntimeLeaseReleased(t, plane)
}

func failedRuntimeResult() agentbridge.Result {
	return agentbridge.Result{Status: agentbridge.ResultFailed, Error: "boom"}
}

func assertRuntimeLeaseReleased(t *testing.T, plane *Plane) {
	t.Helper()
	registry := readRuntimeLeaseRegistry(t, plane.leasePath)
	if len(registry.Leases) != 1 || registry.Leases[0].ReleasedAt == nil {
		t.Fatalf("lease should be released after terminal report: %+v", registry.Leases)
	}
}
