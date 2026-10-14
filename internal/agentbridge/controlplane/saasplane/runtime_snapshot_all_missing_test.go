package saasplane

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRegistersAllProvidersMissingNotEmpty(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newRuntimeBindingPlane(t, fake, nil)

	for _, provider := range []string{"claude", "codex", "openclaw", "cursor"} {
		registerRuntimeForBinding(t, plane, missingProviderRegistration(provider))
	}

	if len(fake.runtimeSnapshots) == 0 {
		t.Fatal("expected runtime snapshots to be posted")
	}
	last := fake.runtimeSnapshots[len(fake.runtimeSnapshots)-1]
	if len(last.Runtimes) != 4 {
		t.Fatalf("final snapshot must carry all four providers, got %+v", last.Runtimes)
	}
	assertRuntimeSnapshotAllMissing(t, last.Runtimes)
}

func missingProviderRegistration(provider string) controlplane.RuntimeRegistration {
	return controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:" + provider,
		Provider:  provider,
		Capabilities: map[string]bool{
			"provider." + provider + ".available": false,
		},
	}
}

func assertRuntimeSnapshotAllMissing(t *testing.T, runtimes []RuntimeSnapshotRecord) {
	t.Helper()
	for _, rt := range runtimes {
		if rt.Availability != "offline" || rt.DetectionState != "missing" {
			t.Fatalf("every undetected provider must be present as missing, got %+v", rt)
		}
	}
}
