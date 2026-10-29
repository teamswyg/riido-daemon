package runtimeactor

import (
	"github.com/teamswyg/riido-contracts/hostintegration"
	"testing"
)

func TestFailedVersionProbeRetainsOnlyVerifiedBinaryIdentity(t *testing.T) {
	previous := Capability{Provider: "codex", Executable: "/test/codex", Version: "codex-cli 1.2.3", DetectedFingerprint: "binary-a", CapabilityFingerprint: "cap-a"}
	for _, changed := range []bool{false, true} {
		observed := previous
		observed.Version = ""
		observed.HealthStatus = hostintegration.ProviderHealthUnknown
		observed.DiagnosticCode = hostintegration.ProviderDiagnosticVersionProbeFailed
		if changed {
			observed.DetectedFingerprint = "binary-b"
		}
		got := retainCapabilityProbeIdentity(previous, observed)
		if !changed && got.Version != previous.Version {
			t.Fatal("same binary lost last known version")
		}
		if changed && got.Version != "" {
			t.Fatal("new binary inherited unverified version")
		}
		if got.DiagnosticCode != observed.DiagnosticCode {
			t.Fatal("version failure was concealed")
		}
	}
}
