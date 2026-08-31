package openclaw

import (
	"context"
	"testing"

	"github.com/teamswyg/riido-contracts/hostintegration"
)

func TestDetectOpenClawAuthExitOneDistinguishesMissingFromIndeterminate(t *testing.T) {
	tests := []struct {
		name       string
		authJSON   string
		available  bool
		health     hostintegration.ProviderHealthStatus
		diagnostic hostintegration.ProviderDiagnosticCode
	}{
		{
			name:       "native CLI auth is indeterminate but executable",
			authJSON:   `{"auth":{"missingProvidersInUse":[],"unusableProfiles":[],"modelRouteIssues":[{"kind":"indeterminate"}]}}`,
			available:  true,
			health:     hostintegration.ProviderHealthDegraded,
			diagnostic: hostintegration.ProviderDiagnosticAuthProbeFailed,
		},
		{
			name:       "provider credential is explicitly missing",
			authJSON:   `{"auth":{"missingProvidersInUse":["anthropic"],"unusableProfiles":[]}}`,
			available:  false,
			health:     hostintegration.ProviderHealthUnavailable,
			diagnostic: hostintegration.ProviderDiagnosticLoginRequired,
		},
		{
			name:       "exit one without an indeterminate route stays unavailable",
			authJSON:   `{"auth":{"missingProvidersInUse":[],"unusableProfiles":[]}}`,
			available:  false,
			health:     hostintegration.ProviderHealthUnavailable,
			diagnostic: hostintegration.ProviderDiagnosticLoginRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := detectExecutable(context.Background(), writeAuthShim(t, tt.authJSON, 1))
			if res.Available != tt.available || res.HealthStatus != tt.health || res.DiagnosticCode != tt.diagnostic {
				t.Fatalf("detect result = %+v", res)
			}
		})
	}
}
