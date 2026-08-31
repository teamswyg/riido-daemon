package supervisor

import (
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func TestProviderHealthObservationIsBoundedForSupportedProviders(t *testing.T) {
	tests := []struct {
		provider   string
		available  bool
		reason     string
		wantHealth hostintegration.ProviderHealthStatus
		wantCode   hostintegration.ProviderDiagnosticCode
	}{
		{"codex", true, "", hostintegration.ProviderHealthHealthy, hostintegration.ProviderDiagnosticNone},
		{"claude", false, "authentication credentials expired", hostintegration.ProviderHealthUnavailable, hostintegration.ProviderDiagnosticLoginRequired},
		{"openclaw", true, "provider login requires attention", hostintegration.ProviderHealthDegraded, hostintegration.ProviderDiagnosticLoginRequired},
		{"openclaw", false, "older than minimum; upgrade openclaw", hostintegration.ProviderHealthUnavailable, hostintegration.ProviderDiagnosticVersionUnsupported},
		{"cursor", false, "cursor-agent executable not found", hostintegration.ProviderHealthUnavailable, hostintegration.ProviderDiagnosticExecutableMissing},
	}
	for _, test := range tests {
		t.Run(test.provider, func(t *testing.T) {
			health, code, _ := providerHealthObservation(runtimeactor.Capability{Provider: test.provider, Available: test.available, Reason: test.reason})
			if health != test.wantHealth || code != test.wantCode {
				t.Fatalf("health=%q code=%q", health, code)
			}
		})
	}
}

func TestProviderHealthUnknownDoesNotProjectRawReason(t *testing.T) {
	secret := "unexpected failure bearer-secret /Users/private/provider"
	health, code, summary := providerHealthObservation(runtimeactor.Capability{Provider: "codex", Reason: secret})
	if health != hostintegration.ProviderHealthUnknown || code != hostintegration.ProviderDiagnosticProbeFailed {
		t.Fatalf("health=%q code=%q", health, code)
	}
	if strings.Contains(summary, "secret") || strings.Contains(summary, "/Users") {
		t.Fatalf("summary leaked raw reason: %q", summary)
	}
}

func TestProviderHealthUsesTypedProbeFailure(t *testing.T) {
	secret := "bearer-secret /Users/private/provider"
	health, code, summary := providerHealthObservation(runtimeactor.Capability{
		Provider:       "codex",
		Available:      true,
		HealthStatus:   hostintegration.ProviderHealthUnknown,
		DiagnosticCode: hostintegration.ProviderDiagnosticAuthProbeFailed,
		Reason:         secret,
	})
	if health != hostintegration.ProviderHealthUnknown || code != hostintegration.ProviderDiagnosticAuthProbeFailed {
		t.Fatalf("health=%q code=%q", health, code)
	}
	if summary != "provider authentication probe did not complete" || strings.Contains(summary, secret) {
		t.Fatalf("summary=%q", summary)
	}
}

func TestRuntimeRegistrationStateChangesWithProviderHealth(t *testing.T) {
	status := runtimeactor.Status{RuntimeID: "daemon:codex", Capabilities: []runtimeactor.Capability{{Provider: "codex", Available: true}}}
	before := runtimeRegistrationState(status)
	status.Capabilities[0].Available = false
	status.Capabilities[0].Reason = "probe failed"
	if after := runtimeRegistrationState(status); after == before {
		t.Fatal("provider health change must trigger runtime re-registration")
	}
}
