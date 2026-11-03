package saasplane

import (
	"testing"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestRuntimeSnapshotKeepsVersionObservationSeparateFromHeartbeat(t *testing.T) {
	const observed = "2026-10-29T00:00:00Z"
	snapshot, _, ok := runtimeSnapshotFromRegistration(controlplane.RuntimeRegistration{RuntimeID: "device:codex", Provider: "codex", CapabilityAttributes: map[string]string{"provider.codex.provider_version": "1.2.3", "provider.codex.provider_version_observed_at": observed}})
	if !ok || snapshot.ProviderVersionObservedAt != observed {
		t.Fatalf("missing version observation: %+v", snapshot)
	}
}
