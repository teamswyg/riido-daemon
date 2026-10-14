package taskdbplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestHeartbeatRefreshesRuntimeLeaseForRunningTask(t *testing.T) {
	path := writeQueuedCodexTaskDB(t)
	plane := newTestPlane(t, path)
	now := time.Date(2026, 5, 25, 3, 0, 0, 0, time.UTC)
	plane.now = func() time.Time { return now }
	registerRuntimeForProvider(t, plane, "runtime-1", "codex", 2, 0)
	claimRuntimeTask(t, plane, "runtime-1")
	before := readRuntimeLeaseRegistry(t, plane.leasePath).Leases[0].LeaseUntil

	now = now.Add(10 * time.Second)
	if err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{
		RuntimeID:      "runtime-1",
		SlotLimit:      2,
		SlotsInUse:     1,
		RunningTaskIDs: []string{codexTaskID},
	}); err != nil {
		t.Fatalf("Heartbeat: %v", err)
	}
	after := readRuntimeLeaseRegistry(t, plane.leasePath).Leases[0].LeaseUntil
	if !after.After(before) {
		t.Fatalf("heartbeat should refresh lease deadline: before=%s after=%s", before, after)
	}
}
