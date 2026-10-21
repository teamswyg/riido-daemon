package agentbridge

import "strings"

func AgentInstructionContract(instruction string) string {
	instruction = strings.TrimSpace(instruction)
	if instruction == "" {
		return ""
	}
	return "Riido agent instruction:\n" + instruction
}

func TelemetryContractInstruction() string {
	return "Riido telemetry contract:\n- " + strings.Join(TelemetryNativeConfigHardRules(), "\n- ")
}

func TelemetryNativeConfigHardRules() []string {
	return []string{
		`Emit structured progress telemetry <riido_log>{"code":1103,"args":{"label":"...","description":"..."}}<end> only when meaningful work changes. Use code 1001 with empty args only for an initial thinking state when no concrete progress can yet be reported.`,
		"Use only Riido progress message codes with these meanings: 1001 thinking, 1101 tool collecting, 1102 collection completed count, 1103 tool running, and 1104 tool completed.",
		"Use args label and description for 1101/1103, label/count/representative_title for 1102, and label/summary for 1104.",
		`Keep label args state-neutral; never include rendered state words like "수집 중", "조회 완료", "실행 중", "실행", or "완료".`,
		"Write description and summary in the user's language as one short sentence of approximately 120 characters describing the actual current or completed action for this task. Preserve concrete task context instead of generic fixed status copy.",
		"Do not repeat the same progress, emit periodic filler, or include raw commands, code, internal function or tool names, debug logs, prompt dumps, secrets, or long answer drafts.",
		"Never invent tool results, success, waiting reasons, or another task being busy. Report only observed work; emit completed progress only after verifying the outcome. If no meaningful change has occurred, emit no new progress.",
		"Use <riido_log> only for progress telemetry, not for final code blocks.",
		"Do not invent new Riido progress codes or argument fields. Keep partial answers and final results separate from progress telemetry.",
	}
}
