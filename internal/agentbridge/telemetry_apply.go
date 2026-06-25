package agentbridge

import "strings"

func ApplyTelemetryContract(provider, prompt, systemPrompt string) (string, string, string) {
	strategy := RuntimeInstructionStrategyForProvider(provider)
	switch strategy.TelemetryPlacement {
	case TelemetryPlacementSystemPrompt, TelemetryPlacementSystemPromptInline:
		return strings.TrimSpace(prompt), appendPromptSection(systemPrompt, TelemetryContractInstruction()), strategy.TelemetryPlacement
	case TelemetryPlacementPrompt:
		return InjectTelemetryContract(prompt), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt
	default:
		return InjectTelemetryContract(prompt), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt
	}
}

func ApplyRuntimeInstructionContract(provider, prompt, systemPrompt, agentInstruction string) (string, string, string, string) {
	strategy := RuntimeInstructionStrategyForProvider(provider)
	agentSection := AgentInstructionContract(agentInstruction)
	interactionSection := AssignmentInteractionContractInstruction()
	switch strategy.AgentInstructionPlacement {
	case TelemetryPlacementSystemPrompt, TelemetryPlacementSystemPromptInline:
		system := appendPromptSection(systemPrompt, agentSection)
		system = appendPromptSection(system, interactionSection)
		system = appendPromptSection(system, TelemetryContractInstruction())
		return strings.TrimSpace(prompt), system, strategy.TelemetryPlacement, placementForSection(agentSection, strategy.AgentInstructionPlacement)
	case TelemetryPlacementPrompt:
		return InjectPromptSections(prompt, agentSection, interactionSection, TelemetryContractInstruction()), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt, placementForSection(agentSection, TelemetryPlacementPrompt)
	default:
		return InjectPromptSections(prompt, agentSection, interactionSection, TelemetryContractInstruction()), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt, placementForSection(agentSection, TelemetryPlacementPrompt)
	}
}
