package saasplane

import (
	"strings"

	"github.com/teamswyg/riido-contracts/hostintegration"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func runtimeModels(in []controlplane.RuntimeModel) []RuntimeModelRecord {
	out := make([]RuntimeModelRecord, 0, len(in))
	for _, model := range in {
		out = append(out, RuntimeModelRecord{
			ModelID:   model.ModelID,
			Label:     model.Label,
			IsDefault: model.IsDefault,
		})
	}
	return out
}

func runtimeProviderHealthStatus(rt controlplane.RuntimeRegistration, provider string) hostintegration.ProviderHealthStatus {
	value := hostintegration.ProviderHealthStatus(strings.TrimSpace(rt.CapabilityAttributes["provider."+provider+".health_status"]))
	if value.Valid() {
		return value
	}
	if available, ok := rt.Capabilities["provider."+provider+".available"]; ok && !available {
		return hostintegration.ProviderHealthUnavailable
	}
	return hostintegration.ProviderHealthHealthy
}

func runtimeProviderDiagnosticCode(rt controlplane.RuntimeRegistration, provider string) hostintegration.ProviderDiagnosticCode {
	value := hostintegration.ProviderDiagnosticCode(strings.TrimSpace(rt.CapabilityAttributes["provider."+provider+".diagnostic_code"]))
	if value.Valid() {
		return value
	}
	return hostintegration.ProviderDiagnosticNone
}

func runtimeProviderDiagnosticSummary(rt controlplane.RuntimeRegistration, provider string) string {
	return strings.TrimSpace(rt.CapabilityAttributes["provider."+provider+".diagnostic_summary"])
}

func runtimeRequiresExperimentalOptIn(rt controlplane.RuntimeRegistration, provider string) bool {
	if len(rt.Capabilities) == 0 {
		return false
	}
	key := "provider." + provider + ".requires_experimental_opt_in"
	return rt.Capabilities[key]
}

func runtimeProviderVersion(rt controlplane.RuntimeRegistration, provider string) string {
	return strings.TrimSpace(rt.CapabilityAttributes["provider."+provider+".provider_version"])
}
