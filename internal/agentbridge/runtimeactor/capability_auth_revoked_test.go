package runtimeactor

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/bridge"
)

func TestConfirmedLogoutBlocksSessionWithAuthenticationReason(t *testing.T) {
	actor, adapter, handle, observed := startPinnedProbeFixture(t)
	observed.Available = false
	observed.HealthStatus = hostintegration.ProviderHealthUnavailable
	observed.DiagnosticCode = hostintegration.ProviderDiagnosticLoginRequired
	adapter.setDetected(observed)
	refreshProbeFixture(t, actor)
	result := expectTaskResult(t, handle.Result(), "confirmed logout did not stop running work")
	if result.Status != agentbridge.ResultBlocked || !strings.Contains(result.Error, "authentication required") || strings.Contains(result.Error, "runtime pin violated") {
		t.Fatalf("logout misclassified as runtime drift: %+v", result)
	}
	if _, err := actor.Submit(context.Background(), bridge.TaskRequest{ID: "new", Provider: "codex"}); !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("logout admitted new work: %v", err)
	}
}

func TestBinaryDriftWinsOverAuthenticationFailure(t *testing.T) {
	for _, diagnostic := range []hostintegration.ProviderDiagnosticCode{
		hostintegration.ProviderDiagnosticLoginRequired,
		hostintegration.ProviderDiagnosticVersionProbeFailed,
	} {
		t.Run(string(diagnostic), func(t *testing.T) {
			actor, adapter, handle, observed := startPinnedProbeFixture(t)
			if err := os.WriteFile(observed.Executable, []byte("replacement provider"), 0o700); err != nil {
				t.Fatal(err)
			}
			observed.Available = false
			observed.HealthStatus = hostintegration.ProviderHealthUnknown
			if diagnostic == hostintegration.ProviderDiagnosticLoginRequired {
				observed.HealthStatus = hostintegration.ProviderHealthUnavailable
			}
			observed.DiagnosticCode = diagnostic
			adapter.setDetected(observed)
			refreshProbeFixture(t, actor)
			result := expectTaskResult(t, handle.Result(), "binary replacement did not fence running work")
			if result.Status != agentbridge.ResultBlocked || !strings.Contains(result.Error, ErrRuntimePinViolated.Error()) {
				t.Fatalf("authentication concealed binary drift: %+v", result)
			}
		})
	}
}
