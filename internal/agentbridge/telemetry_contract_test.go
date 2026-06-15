package agentbridge

import (
	"strings"
	"testing"
)

func TestApplyRuntimeInstructionContractPlacesByProvider(t *testing.T) {
	tests := []struct {
		provider        string
		wantInstruction string
		wantTelemetry   string
		wantPromptHas   []string
		wantSystemHas   []string
		wantPromptExact bool
	}{
		{
			provider:        "claude",
			wantInstruction: TelemetryPlacementSystemPrompt,
			wantTelemetry:   TelemetryPlacementSystemPrompt,
			wantPromptHas:   []string{"do it"},
			wantSystemHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>"},
			wantPromptExact: true,
		},
		{
			provider:        "openclaw",
			wantInstruction: TelemetryPlacementSystemPromptInline,
			wantTelemetry:   TelemetryPlacementSystemPromptInline,
			wantPromptHas:   []string{"do it"},
			wantSystemHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>"},
			wantPromptExact: true,
		},
		{
			provider:        "codex",
			wantInstruction: TelemetryPlacementPrompt,
			wantTelemetry:   TelemetryPlacementPrompt,
			wantPromptHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>", "User task:", "do it"},
		},
		{
			provider:        "cursor",
			wantInstruction: TelemetryPlacementPrompt,
			wantTelemetry:   TelemetryPlacementPrompt,
			wantPromptHas:   []string{"Riido agent instruction:", "act as a PM", "<riido_log>", "User task:", "do it"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			prompt, system, telemetryPlacement, instructionPlacement := ApplyRuntimeInstructionContract(tt.provider, "do it", "", "act as a PM")
			if telemetryPlacement != tt.wantTelemetry || instructionPlacement != tt.wantInstruction {
				t.Fatalf("placements telemetry=%q instruction=%q", telemetryPlacement, instructionPlacement)
			}
			if tt.wantPromptExact && prompt != "do it" {
				t.Fatalf("prompt = %q", prompt)
			}
			for _, want := range tt.wantPromptHas {
				if !strings.Contains(prompt, want) {
					t.Fatalf("prompt missing %q: %q", want, prompt)
				}
			}
			for _, want := range tt.wantSystemHas {
				if !strings.Contains(system, want) {
					t.Fatalf("system missing %q: %q", want, system)
				}
			}
		})
	}
}

func TestRuntimeInstructionStrategiesAreStable(t *testing.T) {
	want := []RuntimeInstructionStrategy{
		{
			Provider:                  "claude",
			AgentInstructionPlacement: TelemetryPlacementSystemPrompt,
			TelemetryPlacement:        TelemetryPlacementSystemPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  "openclaw",
			AgentInstructionPlacement: TelemetryPlacementSystemPromptInline,
			TelemetryPlacement:        TelemetryPlacementSystemPromptInline,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  "codex",
			AgentInstructionPlacement: TelemetryPlacementPrompt,
			TelemetryPlacement:        TelemetryPlacementPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
		{
			Provider:                  "cursor",
			AgentInstructionPlacement: TelemetryPlacementPrompt,
			TelemetryPlacement:        TelemetryPlacementPrompt,
			EffectivenessGate:         "opt-in-real-provider-probe",
		},
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

func TestBuildInstructionEffectivenessProbePlacesMarkerByProvider(t *testing.T) {
	for _, strategy := range RuntimeInstructionStrategies() {
		t.Run(strategy.Provider, func(t *testing.T) {
			probe := BuildInstructionEffectivenessProbe(strategy.Provider)
			if probe.AgentInstructionPlacement != strategy.AgentInstructionPlacement {
				t.Fatalf("instruction placement = %q, want %q", probe.AgentInstructionPlacement, strategy.AgentInstructionPlacement)
			}
			if probe.TelemetryPlacement != strategy.TelemetryPlacement {
				t.Fatalf("telemetry placement = %q, want %q", probe.TelemetryPlacement, strategy.TelemetryPlacement)
			}
			promptHasMarker := strings.Contains(probe.Prompt, probe.ExpectedMarker)
			systemHasMarker := strings.Contains(probe.SystemPrompt, probe.ExpectedMarker)
			switch strategy.AgentInstructionPlacement {
			case TelemetryPlacementPrompt:
				if !promptHasMarker || systemHasMarker {
					t.Fatalf("marker location prompt=%t system=%t probe=%+v", promptHasMarker, systemHasMarker, probe)
				}
			case TelemetryPlacementSystemPrompt, TelemetryPlacementSystemPromptInline:
				if promptHasMarker || !systemHasMarker {
					t.Fatalf("marker location prompt=%t system=%t probe=%+v", promptHasMarker, systemHasMarker, probe)
				}
			default:
				t.Fatalf("unknown placement %q", strategy.AgentInstructionPlacement)
			}
			if err := ValidateInstructionEffectivenessOutput(probe, "ok\n"+probe.ExpectedMarker+"\n"); err != nil {
				t.Fatalf("valid marker rejected: %v", err)
			}
			if err := ValidateInstructionEffectivenessOutput(probe, "ok"); err == nil {
				t.Fatal("missing marker accepted")
			}
		})
	}
}

func TestRuntimeInstructionStrategyForUnknownProviderDefaultsToPrompt(t *testing.T) {
	strategy := RuntimeInstructionStrategyForProvider("future-provider")
	if strategy.Provider != "future-provider" {
		t.Fatalf("provider = %q", strategy.Provider)
	}
	if strategy.AgentInstructionPlacement != TelemetryPlacementPrompt || strategy.TelemetryPlacement != TelemetryPlacementPrompt {
		t.Fatalf("unknown provider strategy = %+v", strategy)
	}
}

func TestInjectTelemetryContractIsIdempotent(t *testing.T) {
	first := InjectTelemetryContract("do it")
	second := InjectTelemetryContract(first)
	if second != first {
		t.Fatalf("telemetry contract duplicated:\nfirst=%q\nsecond=%q", first, second)
	}
}
