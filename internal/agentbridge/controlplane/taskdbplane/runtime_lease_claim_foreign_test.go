package taskdbplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/task"
)

func TestClaimTaskSkipsActiveLeaseOwnedByAnotherRuntime(t *testing.T) {
	path := writeQueuedCodexTaskDB(t)
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)
	now := leaseClaimNow().Add(2 * time.Hour)
	lease := RuntimeLeaseRecord{
		LeaseID:               "runtime-lease:codex-task:7",
		TaskID:                codexTaskID,
		RuntimeID:             "runtime-other",
		CapabilityFingerprint: "runtime-1-fp",
		ClaimedAt:             now,
		LeaseUntil:            now.Add(time.Hour),
		FencingToken:          7,
	}
	writeRuntimeLease(t, plane, lease)

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req != nil {
		t.Fatalf("task with active foreign lease should not be claimed: %+v", req)
	}
	assertTaskState(t, path, codexTaskID, task.StateQueued)
	got := assertSingleLease(t, readRuntimeLeaseRegistry(t, plane.leasePath))
	if got.FencingToken != 7 {
		t.Fatalf("foreign lease should remain unchanged: %+v", got)
	}
}
