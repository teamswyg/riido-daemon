package agentbridge

// RuntimeInstructionStrategy is the daemon-owned provider placement decision
// for assignment-created agent instructions.
type RuntimeInstructionStrategy struct {
	Provider                  string
	AgentInstructionPlacement string
	TelemetryPlacement        string
	EffectivenessGate         string
}

// InstructionEffectivenessProbe is the provider-neutral probe payload used by
// optional real-provider checks to verify that the selected placement is obeyed.
type InstructionEffectivenessProbe struct {
	Provider                  string
	Prompt                    string
	SystemPrompt              string
	ExpectedMarker            string
	AgentInstructionPlacement string
	TelemetryPlacement        string
}

// TelemetryParser extracts Riido control-layer telemetry tags from provider
// text deltas. It is provider-neutral and owned by the session actor.
type TelemetryParser struct {
	buf string
}

func NewTelemetryParser() *TelemetryParser {
	return &TelemetryParser{}
}
