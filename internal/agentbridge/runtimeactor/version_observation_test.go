package runtimeactor

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestVersionObservationSurvivesFailedProbe(t *testing.T) {
	now := time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC)
	previous, err := buildRuntimeCapability("runtime", "codex", agentbridge.DetectResult{Available: true, Version: "codex-cli 1.2.3", Executable: "/missing/codex"}, "test", now)
	if err != nil {
		t.Fatal(err)
	}
	failed, err := buildRuntimeCapability("runtime", "codex", agentbridge.DetectResult{Available: true, Executable: "/missing/codex", HealthStatus: hostintegration.ProviderHealthUnknown, DiagnosticCode: hostintegration.ProviderDiagnosticVersionProbeFailed}, "test", now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	got := retainCapabilityProbeIdentity(previous, failed)
	if got.Version != previous.Version || !got.VersionObservedAt.Equal(now) {
		t.Fatalf("lost original observation: %+v", got)
	}
	if !failed.VersionObservedAt.IsZero() {
		t.Fatal("failed probe manufactured observation time")
	}
	failed.DetectedFingerprint = "new-binary"
	if got := retainCapabilityProbeIdentity(previous, failed); !got.VersionObservedAt.IsZero() {
		t.Fatal("new binary retained old observation")
	}
}
