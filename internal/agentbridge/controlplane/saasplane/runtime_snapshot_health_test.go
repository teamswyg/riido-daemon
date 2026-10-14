package saasplane

import (
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func TestRuntimeSnapshotProjectsBoundedProviderHealth(t *testing.T) {
	prefix := "provider.codex."
	got, _, ok := runtimeSnapshotFromRegistration(controlplane.RuntimeRegistration{
		RuntimeID:    "daemon:codex",
		Provider:     "codex",
		Capabilities: map[string]bool{prefix + "available": false},
		CapabilityAttributes: map[string]string{
			prefix + "health_status":      "unknown",
			prefix + "diagnostic_code":    "probe-failed",
			prefix + "diagnostic_summary": "provider probe did not complete",
		},
	})
	if !ok || got.HealthStatus != hostintegration.ProviderHealthUnknown || got.DiagnosticCode != hostintegration.ProviderDiagnosticProbeFailed {
		t.Fatalf("runtime snapshot = %+v", got)
	}
	if got.Availability != "offline" || got.DetectionState != "error" {
		t.Fatalf("runtime availability = %+v", got)
	}
}
