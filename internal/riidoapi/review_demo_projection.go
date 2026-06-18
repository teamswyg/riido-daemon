package riidoapi

import "github.com/teamswyg/riido-daemon/internal/hostintegration"

func reviewDemoResponseFromMode(mode hostintegration.ReviewDemoMode) ReviewDemoResponse {
	surfaces := make([]string, 0, len(mode.Surfaces))
	for _, surface := range mode.Surfaces {
		surfaces = append(surfaces, string(surface))
	}
	providerStatusMode := "real-status"
	if mode.Enabled {
		providerStatusMode = "synthetic-preview"
	}
	return ReviewDemoResponse{
		SchemaVersion:            ReviewDemoSchemaVersion,
		DistributionChannel:      string(mode.Channel),
		Enabled:                  mode.Enabled,
		Surfaces:                 surfaces,
		ProviderStatusMode:       providerStatusMode,
		ProviderExecutionAllowed: mode.ProviderExecutionAllowed,
		TelemetrySyncAllowed:     mode.TelemetrySyncAllowed,
		LocalOnly:                true,
	}
}
