package runtimeactor

import "github.com/teamswyg/riido-contracts/hostintegration"

// retainCapabilityProbeIdentity prevents incomplete authentication observations
// from being mistaken for a provider upgrade. It never conceals an observed
// executable, binary fingerprint, or version change.
func retainCapabilityProbeIdentity(previous, observed Capability) Capability {
	if observed.HealthStatus != hostintegration.ProviderHealthUnknown ||
		(observed.DiagnosticCode != hostintegration.ProviderDiagnosticAuthProbeFailed && observed.DiagnosticCode != hostintegration.ProviderDiagnosticProbeFailed && observed.DiagnosticCode != hostintegration.ProviderDiagnosticVersionProbeFailed) ||
		previous.Provider != observed.Provider || previous.Executable != observed.Executable ||
		previous.DetectedFingerprint != observed.DetectedFingerprint ||
		(observed.Version != "" && previous.Version != observed.Version) || previous.CapabilityFingerprint == "" {
		return observed
	}
	retained := previous
	if observed.Version != "" {
		retained.VersionObservedAt = observed.VersionObservedAt
	}
	retained.Available = observed.Available
	retained.HealthStatus = observed.HealthStatus
	retained.DiagnosticCode = observed.DiagnosticCode
	retained.Reason = observed.Reason
	return retained
}
