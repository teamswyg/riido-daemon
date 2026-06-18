package saasplane

import (
	"context"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRegistersRuntimeSnapshotWithDeviceCredential(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	startedAt := time.Now().Add(-2 * time.Minute).UTC()
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
		Profile:      "development",
		AppVersion:   "v0.0.13",
		PID:          4321,
		StartedAt:    startedAt,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	err = plane.RegisterRuntime(context.Background(), runtimeSnapshotRegistration())
	if err != nil {
		t.Fatalf("RegisterRuntime: %v", err)
	}
	if len(fake.runtimeSnapshots) != 1 {
		t.Fatalf("runtime snapshots = %+v", fake.runtimeSnapshots)
	}
	assertRuntimeSnapshotIdentity(t, fake.runtimeSnapshots[0], startedAt)
	assertRuntimeSnapshotRuntime(t, fake.runtimeSnapshots[0].Runtimes)
}

func runtimeSnapshotRegistration() controlplane.RuntimeRegistration {
	return controlplane.RuntimeRegistration{
		DaemonID:   "daemon-1",
		RuntimeID:  "daemon-1:codex",
		Provider:   "codex",
		DeviceName: "주윤의 MacBook",
		Models: []controlplane.RuntimeModel{
			{ModelID: "gpt-5.5", Label: "gpt-5.5", IsDefault: true},
		},
		Capabilities: map[string]bool{
			"provider.codex.requires_experimental_opt_in": true,
		},
		CapabilityAttributes: map[string]string{
			"provider.codex.provider_version": "codex-cli 0.133.0",
		},
	}
}
