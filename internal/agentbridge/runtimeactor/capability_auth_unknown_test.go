package runtimeactor

import (
	"context"
	"errors"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestIncompleteProbePreservesSessionButRefusesAdmission(t *testing.T) {
	for _, diagnostic := range []hostintegration.ProviderDiagnosticCode{
		hostintegration.ProviderDiagnosticAuthProbeFailed,
		hostintegration.ProviderDiagnosticVersionProbeFailed,
	} {
		t.Run(string(diagnostic), func(t *testing.T) {
			actor, adapter, handle, verified := startPinnedProbeFixture(t)
			original := actorStatusCapabilities(t, actor)[0]
			unknown := verified
			unknown.Version = ""
			unknown.HealthStatus, unknown.DiagnosticCode = hostintegration.ProviderHealthUnknown, diagnostic
			adapter.setDetected(unknown)
			observed := refreshProbeFixture(t, actor)
			if observed.Available || observed.CapabilityFingerprint != original.CapabilityFingerprint {
				t.Fatalf("incomplete observation changed identity or admitted work: %+v", observed)
			}
			if _, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "new", Provider: "codex"}); !errors.Is(err, ErrProviderUnavailable) {
				t.Fatalf("new work admitted: %v", err)
			}
			select {
			case result := <-handle.Result():
				t.Fatalf("incomplete probe stopped running work: %+v", result)
			default:
			}
			adapter.setDetected(verified)
			recovered := refreshProbeFixture(t, actor)
			if !recovered.Available || recovered.CapabilityFingerprint != original.CapabilityFingerprint {
				t.Fatal("verified recovery did not restore admission with the original identity")
			}
		})
	}
}

func TestAdapterProbeErrorRefusesAdmissionWithoutStoppingSession(t *testing.T) {
	actor, adapter, handle, _ := startPinnedProbeFixture(t)
	adapter.failDetection()
	if err := actor.RefreshCapability(context.Background(), "codex"); err == nil {
		t.Fatal("probe error was not reported")
	}
	if _, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "new", Provider: "codex"}); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("failed probe admitted new work: %v", err)
	}
	select {
	case result := <-handle.Result():
		t.Fatalf("probe error stopped running work: %+v", result)
	default:
	}
}
