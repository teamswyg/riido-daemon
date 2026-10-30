package runtimeactor

import "github.com/teamswyg/riido-contracts/hostintegration"

// retainCapabilityProbeIdentity separates authentication/probe availability
// from a provider upgrade. It never conceals an observed
// executable, binary fingerprint, or version change.
func retainCapabilityProbeIdentity(previous, observed Capability) Capability {
	if !capabilityIdentityUnchangedByProbe(observed) ||
		previous.Provider != observed.Provider || previous.Executable != observed.Executable ||
		previous.DetectedFingerprint != observed.DetectedFingerprint ||
		(observed.Version != "" && previous.Version != observed.Version) || previous.CapabilityFingerprint == "" {
		return observed
	}
	retained := previous
	if observed.Version != "" {
		retained.VersionObservedAt = observed.VersionObservedAt
	}
	retained.Available = false
	retained.HealthStatus = observed.HealthStatus
	retained.DiagnosticCode = observed.DiagnosticCode
	retained.Reason = observed.Reason
	return retained
}

func capabilityIdentityUnchangedByProbe(observed Capability) bool {
	if observed.HealthStatus == hostintegration.ProviderHealthUnavailable {
		return observed.DiagnosticCode == hostintegration.ProviderDiagnosticLoginRequired
	}
	if observed.HealthStatus != hostintegration.ProviderHealthUnknown {
		return false
	}
	switch observed.DiagnosticCode {
	case hostintegration.ProviderDiagnosticAuthProbeFailed,
		hostintegration.ProviderDiagnosticVersionProbeFailed,
		hostintegration.ProviderDiagnosticProbeFailed:
		return true
	default:
		return false
	}
}
