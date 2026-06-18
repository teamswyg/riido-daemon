package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestClaimTaskRequeuesExpiredRunningLeaseForHandoff(t *testing.T) {
	path := writeTaskDB(t, runningCodexTaskDB())
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-new", "codex", 2, 0)
	writeRuntimeLease(t, plane, runtimeLeaseFixture(
		codexTaskID,
		"runtime-old",
		"runtime-old-fp",
		3,
		leaseClaimNow(),
	))

	req, err := plane.ClaimTask(context.Background(), "runtime-new")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.ID != codexTaskID || req.Metadata[controlplane.MetadataRuntimeFencingToken] != "4" {
		t.Fatalf("expired running lease should be handed off and claimed with token 4, got %+v", req)
	}
	db := loadTaskDB(t, path)
	if got := mustFindTask(t, db, codexTaskID).State; got != task.StateClaimed {
		t.Fatalf("task should be claimed by new runtime after handoff, got %s", got)
	}
	assertTransition(t, db, ir.EventBlockerRaised)
	assertTransition(t, db, ir.EventBlockerResolvedRequeue)
	assertTransition(t, db, ir.EventTaskClaimed)
	lease := assertSingleLease(t, readRuntimeLeaseRegistry(t, plane.leasePath))
	if lease.RuntimeID != "runtime-new" || lease.FencingToken != 4 {
		t.Fatalf("new lease mismatch after handoff: %+v", lease)
	}
}
