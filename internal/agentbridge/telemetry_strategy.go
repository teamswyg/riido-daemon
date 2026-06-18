package agentbridge

import providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"

func RuntimeInstructionStrategies() []RuntimeInstructionStrategy {
	strategies := []RuntimeInstructionStrategy{
		{
			Provider:                  string(providercatalog.KindClaude),
			AgentInstructionPlacement: TelemetryPlacementSystemPrompt,
			TelemetryPlacement:        TelemetryPlacementSystemPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  string(providercatalog.KindOpenClaw),
			AgentInstructionPlacement: TelemetryPlacementSystemPromptInline,
			TelemetryPlacement:        TelemetryPlacementSystemPromptInline,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  string(providercatalog.KindCodex),
			AgentInstructionPlacement: TelemetryPlacementPrompt,
			TelemetryPlacement:        TelemetryPlacementPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  string(providercatalog.KindCursor),
			AgentInstructionPlacement: TelemetryPlacementPrompt,
			TelemetryPlacement:        TelemetryPlacementPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
	}
	out := make([]RuntimeInstructionStrategy, len(strategies))
	copy(out, strategies)
	return out
}

func RuntimeInstructionStrategyForProvider(provider string) RuntimeInstructionStrategy {
	provider = string(providercatalog.Normalize(provider))
	if provider == "" {
		provider = "default"
	}
	for _, strategy := range RuntimeInstructionStrategies() {
		if strategy.Provider == provider {
			return strategy
		}
	}
	return RuntimeInstructionStrategy{
		Provider:                  provider,
		AgentInstructionPlacement: TelemetryPlacementPrompt,
		TelemetryPlacement:        TelemetryPlacementPrompt,
		EffectivenessGate:         "opt-in-real-provider-probe",
	}
}
