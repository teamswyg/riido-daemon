package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func TestClaimTaskFailsExpiredClaimedLease(t *testing.T) {
	path := writeTaskDB(t, claimedTaskDB())
	plane := newTestPlane(t, path)
	writeRuntimeLease(t, plane, runtimeLeaseFixture(
		"task-1",
		"runtime-old",
		"runtime-old-fp",
		5,
		leaseClaimNow(),
	))

	req, err := plane.ClaimTask(context.Background(), "runtime-new")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req != nil {
		t.Fatalf("expired claimed lease should fail task rather than hand off: %+v", req)
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, "task-1").State; got != task.StateFailed {
		t.Fatalf("task should fail after expired Claimed lease, got %s", got)
	}
	assertTransition(t, db, ir.EventTaskFailed)
	lease := assertSingleLease(t, readRuntimeLeaseRegistry(t, plane.leasePath))
	if lease.ReleasedAt == nil {
		t.Fatalf("expired claimed lease should be released: %+v", lease)
	}
}
