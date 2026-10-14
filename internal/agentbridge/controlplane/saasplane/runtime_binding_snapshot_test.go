package saasplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneHeartbeatRefreshesAggregatedRuntimeSnapshot(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	startedAt := runtimeBindingStartedAt()
	plane := newRuntimeBindingPlane(t, fake, func(cfg *Config) {
		cfg.Profile = "development"
		cfg.AppVersion = "v0.0.13"
		cfg.PID = 8765
		cfg.StartedAt = startedAt
	})
	registerRuntimeForBinding(t, plane, runtimeSnapshotRegistration())
	registerRuntimeForBinding(t, plane, cursorRuntimeRegistration())
	if len(fake.runtimeSnapshots) != 2 {
		t.Fatalf("registration snapshots = %+v", fake.runtimeSnapshots)
	}

	err := plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{RuntimeID: "daemon-1:codex"})
	if err != nil {
		t.Fatalf("Heartbeat codex: %v", err)
	}
	if len(fake.runtimeSnapshots) != 3 {
		t.Fatalf("heartbeat should append one aggregated snapshot, got %+v", fake.runtimeSnapshots)
	}
	assertRuntimeBindingSnapshot(t, fake.runtimeSnapshots[2], startedAt)

	err = plane.Heartbeat(context.Background(), controlplane.RuntimeHeartbeat{RuntimeID: "daemon-1:cursor"})
	if err != nil {
		t.Fatalf("Heartbeat cursor: %v", err)
	}
	if len(fake.runtimeSnapshots) != 3 {
		t.Fatalf("same heartbeat window should not create per-runtime snapshot fanout: %+v", fake.runtimeSnapshots)
	}
}
