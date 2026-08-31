package supervisor

import (
	"strings"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/runtimeactor"
)

func providerHealthObservation(capability runtimeactor.Capability) (
	hostintegration.ProviderHealthStatus,
	hostintegration.ProviderDiagnosticCode,
	string,
) {
	reason := strings.ToLower(strings.TrimSpace(capability.Reason))
	if capability.Available && reason == "" {
		return hostintegration.ProviderHealthHealthy, hostintegration.ProviderDiagnosticNone, ""
	}
	switch {
	case capability.Available && strings.Contains(reason, "requires attention"):
		return hostintegration.ProviderHealthDegraded, hostintegration.ProviderDiagnosticLoginRequired, "provider login is required"
	case capability.Available:
		return hostintegration.ProviderHealthUnknown, hostintegration.ProviderDiagnosticProbeFailed, "provider probe did not complete"
	case strings.Contains(reason, "not found"):
		return hostintegration.ProviderHealthUnavailable, hostintegration.ProviderDiagnosticExecutableMissing, "provider executable is not available"
	case strings.Contains(reason, "login") || strings.Contains(reason, "authentication") || strings.Contains(reason, "credentials"):
		return hostintegration.ProviderHealthUnavailable, hostintegration.ProviderDiagnosticLoginRequired, "provider login is required"
	case strings.Contains(reason, "older than minimum") || strings.Contains(reason, "upgrade") || strings.Contains(reason, "unsupported"):
		return hostintegration.ProviderHealthUnavailable, hostintegration.ProviderDiagnosticVersionUnsupported, "provider version is not supported"
	default:
		return hostintegration.ProviderHealthUnknown, hostintegration.ProviderDiagnosticProbeFailed, "provider probe did not complete"
	}
}
