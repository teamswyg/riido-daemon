package saasplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRegistersUnavailableRuntimeSnapshotAsOffline(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	plane := newRuntimeBindingPlane(t, fake, nil)

	err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:   "daemon-1",
		RuntimeID:  "daemon-1:openclaw",
		Provider:   "openclaw",
		DeviceName: "주윤의 MacBook",
		Capabilities: map[string]bool{
			"provider.openclaw.available":                    false,
			"provider.openclaw.requires_experimental_opt_in": true,
		},
	})
	if err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}
	if len(fake.runtimeSnapshots) != 1 {
		t.Fatalf("runtime snapshots = %+v", fake.runtimeSnapshots)
	}
	runtime := fake.runtimeSnapshots[0].Runtimes[0]
	if runtime.RuntimeID != "daemon-1:openclaw" ||
		runtime.Kind != "openclaw" ||
		runtime.Availability != "offline" ||
		runtime.DetectionState != "missing" ||
		!runtime.RequiresExperimentalOptIn {
		t.Fatalf("snapshot runtime = %+v", runtime)
	}
}
