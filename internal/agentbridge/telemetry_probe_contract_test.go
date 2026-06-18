package agentbridge

import (
	"strings"
	"testing"
)

func TestBuildInstructionEffectivenessProbePlacesMarkerByProvider(t *testing.T) {
	for _, strategy := range RuntimeInstructionStrategies() {
		t.Run(strategy.Provider, func(t *testing.T) {
			probe := BuildInstructionEffectivenessProbe(strategy.Provider)
			assertProbePlacement(t, probe, strategy)
			assertProbeMarkerLocation(t, probe, strategy.AgentInstructionPlacement)
			assertProbeValidation(t, probe)
		})
	}
}

func assertProbePlacement(t *testing.T, probe InstructionEffectivenessProbe, strategy RuntimeInstructionStrategy) {
	t.Helper()
	if probe.AgentInstructionPlacement != strategy.AgentInstructionPlacement {
		t.Fatalf("instruction placement = %q, want %q", probe.AgentInstructionPlacement, strategy.AgentInstructionPlacement)
	}
	if probe.TelemetryPlacement != strategy.TelemetryPlacement {
		t.Fatalf("telemetry placement = %q, want %q", probe.TelemetryPlacement, strategy.TelemetryPlacement)
	}
}

func assertProbeMarkerLocation(t *testing.T, probe InstructionEffectivenessProbe, placement string) {
	t.Helper()
	promptHasMarker := strings.Contains(probe.Prompt, probe.ExpectedMarker)
	systemHasMarker := strings.Contains(probe.SystemPrompt, probe.ExpectedMarker)
	switch placement {
	case TelemetryPlacementPrompt:
		if !promptHasMarker || systemHasMarker {
			t.Fatalf("marker location prompt=%t system=%t probe=%+v", promptHasMarker, systemHasMarker, probe)
		}
	case TelemetryPlacementSystemPrompt, TelemetryPlacementSystemPromptInline:
		if promptHasMarker || !systemHasMarker {
			t.Fatalf("marker location prompt=%t system=%t probe=%+v", promptHasMarker, systemHasMarker, probe)
		}
	default:
		t.Fatalf("unknown placement %q", placement)
	}
}

func assertProbeValidation(t *testing.T, probe InstructionEffectivenessProbe) {
	t.Helper()
	if err := ValidateInstructionEffectivenessOutput(probe, "ok\n"+probe.ExpectedMarker+"\n"); err != nil {
		t.Fatalf("valid marker rejected: %v", err)
	}
	if err := ValidateInstructionEffectivenessOutput(probe, "ok"); err == nil {
		t.Fatal("missing marker accepted")
	}
}
