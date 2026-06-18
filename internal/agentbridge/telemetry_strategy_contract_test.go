package agentbridge

import "testing"

func TestRuntimeInstructionStrategiesAreStable(t *testing.T) {
	want := []RuntimeInstructionStrategy{
		expectedRuntimeInstructionStrategy("claude", TelemetryPlacementSystemPrompt),
		expectedRuntimeInstructionStrategy("openclaw", TelemetryPlacementSystemPromptInline),
		expectedRuntimeInstructionStrategy("codex", TelemetryPlacementPrompt),
		expectedRuntimeInstructionStrategy("cursor", TelemetryPlacementPrompt),
	}
	got := RuntimeInstructionStrategies()
	if len(got) != len(want) {
		t.Fatalf("strategy count = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("strategy[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestRuntimeInstructionStrategyForUnknownProviderDefaultsToPrompt(t *testing.T) {
	strategy := RuntimeInstructionStrategyForProvider("future-provider")
	if strategy.Provider != "future-provider" {
		t.Fatalf("provider = %q", strategy.Provider)
	}
	if strategy.AgentInstructionPlacement != TelemetryPlacementPrompt ||
		strategy.TelemetryPlacement != TelemetryPlacementPrompt {
		t.Fatalf("unknown provider strategy = %+v", strategy)
	}
}

func expectedRuntimeInstructionStrategy(provider, placement string) RuntimeInstructionStrategy {
	return RuntimeInstructionStrategy{
		Provider:                  provider,
		AgentInstructionPlacement: placement,
		TelemetryPlacement:        placement,
		EffectivenessGate:         "opt-in-real-provider-probe",
	}
}
