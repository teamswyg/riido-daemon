package saasplane

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestPlaneRegistersUnavailableRuntimeSnapshotAsOffline(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	err = plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
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

// Each registration must post the full accumulated provider set so the
// control-plane device projection always shows every known runtime — an
// undetected provider stays present as detection_state=missing rather than
// being dropped to an empty list.
func TestPlaneRegisterPostsFullProviderSetIncludingMissing(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:codex",
		Provider:  "codex",
	}); err != nil {
		t.Fatalf("RegisterRuntime codex: %v", err)
	}
	if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
		DaemonID:  "daemon-1",
		RuntimeID: "daemon-1:claude",
		Provider:  "claude",
		Capabilities: map[string]bool{
			"provider.claude.available": false,
		},
	}); err != nil {
		t.Fatalf("RegisterRuntime claude: %v", err)
	}

	if len(fake.runtimeSnapshots) != 2 {
		t.Fatalf("want one post per register, got %+v", fake.runtimeSnapshots)
	}
	last := fake.runtimeSnapshots[1]
	if len(last.Runtimes) != 2 {
		t.Fatalf("second register must post the full accumulated set, got %+v", last.Runtimes)
	}
	if last.Runtimes[0].RuntimeID != "daemon-1:claude" || last.Runtimes[1].RuntimeID != "daemon-1:codex" {
		t.Fatalf("posted set must be sorted by runtime id, got %+v", last.Runtimes)
	}
	claude := last.Runtimes[0]
	if claude.Availability != "offline" || claude.DetectionState != "missing" {
		t.Fatalf("undetected claude must remain present as missing, got %+v", claude)
	}
	codex := last.Runtimes[1]
	if codex.Availability != "online" || codex.DetectionState != "detected" {
		t.Fatalf("detected codex facts must survive in the full set, got %+v", codex)
	}
}

// Mirrors the daemon: all four provider runtimes register (the supervisor
// always builds claude/codex/openclaw/cursor actors), every one undetected.
// The final posted snapshot must contain four objects, each missing — never
// an empty array.
func TestPlaneRegistersAllProvidersMissingNotEmpty(t *testing.T) {
	fake := newFakeAssignmentServer(t)
	fake.deviceID = "device-1"
	fake.deviceSecret = "rdev-secret"
	plane, err := New(Config{
		BaseURL:      fake.URL(),
		DaemonID:     "daemon-1",
		DeviceID:     "device-1",
		DeviceSecret: "rdev-secret",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer plane.Close()

	for _, provider := range []string{"claude", "codex", "openclaw", "cursor"} {
		if err := plane.RegisterRuntime(context.Background(), controlplane.RuntimeRegistration{
			DaemonID:  "daemon-1",
			RuntimeID: "daemon-1:" + provider,
			Provider:  provider,
			Capabilities: map[string]bool{
				"provider." + provider + ".available": false,
			},
		}); err != nil {
			t.Fatalf("RegisterRuntime %s: %v", provider, err)
		}
	}

	if len(fake.runtimeSnapshots) == 0 {
		t.Fatal("expected runtime snapshots to be posted")
	}
	last := fake.runtimeSnapshots[len(fake.runtimeSnapshots)-1]
	if len(last.Runtimes) != 4 {
		t.Fatalf("final snapshot must carry all four providers, got %+v", last.Runtimes)
	}
	for _, rt := range last.Runtimes {
		if rt.Availability != "offline" || rt.DetectionState != "missing" {
			t.Fatalf("every undetected provider must be present as missing, got %+v", rt)
		}
	}
}
