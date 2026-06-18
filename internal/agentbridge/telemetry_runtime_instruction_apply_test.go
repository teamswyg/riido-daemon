package agentbridge

import (
	"strings"
	"testing"
)

func TestApplyRuntimeInstructionContractPlacesByProvider(t *testing.T) {
	for _, tt := range runtimeInstructionCases() {
		t.Run(tt.provider, func(t *testing.T) {
			prompt, system, telemetryPlacement, instructionPlacement := ApplyRuntimeInstructionContract(
				tt.provider,
				"do it",
				"",
				"act as a PM",
			)
			if telemetryPlacement != tt.wantTelemetry || instructionPlacement != tt.wantInstruction {
				t.Fatalf("placements telemetry=%q instruction=%q", telemetryPlacement, instructionPlacement)
			}
			assertPromptContract(t, prompt, tt)
			assertSystemPromptContract(t, system, tt.wantSystemHas)
		})
	}
}

func assertPromptContract(t *testing.T, prompt string, tt runtimeInstructionCase) {
	t.Helper()
	if tt.wantPromptExact && prompt != "do it" {
		t.Fatalf("prompt = %q", prompt)
	}
	for _, want := range tt.wantPromptHas {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %q", want, prompt)
		}
	}
}

func assertSystemPromptContract(t *testing.T, system string, wantSystemHas []string) {
	t.Helper()
	for _, want := range wantSystemHas {
		if !strings.Contains(system, want) {
			t.Fatalf("system missing %q: %q", want, system)
		}
	}
}
