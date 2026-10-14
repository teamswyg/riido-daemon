package saasplane

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRegisterPostsFullProviderSetIncludingMissing(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newRuntimeBindingPlane(t, fake, nil)

	registerRuntimeForBinding(t, plane, controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:codex",
		Provider:  "codex",
	})
	registerRuntimeForBinding(t, plane, controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:claude",
		Provider:  "claude",
		Capabilities: map[string]bool{
			"provider.claude.available": false,
		},
	})

	if len(fake.runtimeSnapshots) != 2 {
		t.Fatalf("want one post per register, got %+v", fake.runtimeSnapshots)
	}
	assertRuntimeSnapshotFullSet(t, fake.runtimeSnapshots[1])
}

func assertRuntimeSnapshotFullSet(t *testing.T, snapshot DeviceRuntimeSnapshotSyncRequest) {
	t.Helper()
	if len(snapshot.Runtimes) != 2 {
		t.Fatalf("second register must post the full accumulated set, got %+v", snapshot.Runtimes)
	}
	if snapshot.Runtimes[0].RuntimeID != "daemon-1:claude" ||
		snapshot.Runtimes[1].RuntimeID != "daemon-1:codex" {
		t.Fatalf("posted set must be sorted by runtime id, got %+v", snapshot.Runtimes)
	}
	claude := snapshot.Runtimes[0]
	if claude.Availability != "offline" || claude.DetectionState != "missing" {
		t.Fatalf("undetected claude must remain present as missing, got %+v", claude)
	}
	codex := snapshot.Runtimes[1]
	if codex.Availability != "online" || codex.DetectionState != "detected" {
		t.Fatalf("detected codex facts must survive in the full set, got %+v", codex)
	}
}
