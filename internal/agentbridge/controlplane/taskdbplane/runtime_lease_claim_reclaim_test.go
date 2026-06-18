package taskdbplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestClaimTaskReclaimsExpiredLeaseWithIncrementedFencingToken(t *testing.T) {
	path := writeQueuedCodexTaskDB(t)
	plane := newTestPlane(t, path)
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)
	writeRuntimeLease(t, plane, runtimeLeaseFixture(
		codexTaskID,
		"runtime-other",
		"runtime-other-fp",
		7,
		leaseClaimNow(),
	))

	req, err := plane.ClaimTask(context.Background(), "runtime-1")
	if err != nil {
		t.Fatalf("ClaimTask returned error: %v", err)
	}
	if req == nil || req.Metadata[controlplane.MetadataRuntimeFencingToken] != "8" {
		t.Fatalf("expired lease should be reclaimed with token 8, got %+v", req)
	}
	lease := assertSingleLease(t, readRuntimeLeaseRegistry(t, plane.leasePath))
	if lease.RuntimeID != "runtime-1" || lease.FencingToken != 8 {
		t.Fatalf("reclaimed lease mismatch: %+v", lease)
	}
}
