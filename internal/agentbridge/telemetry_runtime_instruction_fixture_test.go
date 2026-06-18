package agentbridge

type runtimeInstructionCase struct {
	provider        string
	wantInstruction string
	wantTelemetry   string
	wantPromptHas   []string
	wantSystemHas   []string
	wantPromptExact bool
}

func runtimeInstructionCases() []runtimeInstructionCase {
	return []runtimeInstructionCase{
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
}
