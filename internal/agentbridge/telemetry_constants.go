package agentbridge

const (
	telemetryLogStart = "<riido_log>"
	telemetryLogEnd   = "<end>"

	// MetadataTelemetryContract records where the Riido telemetry
	// contract was placed for this task. The supervisor uses its
	// presence to mirror the contract into provider-native config.
	MetadataTelemetryContract = "riido_telemetry_contract"
	MetadataAgentInstruction  = "riido_agent_instruction"

	TelemetryPlacementPrompt             = "prompt"
	TelemetryPlacementSystemPrompt       = "system-prompt"
	TelemetryPlacementSystemPromptInline = "system-prompt-inline"
)
