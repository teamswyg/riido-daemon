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
	return `Riido telemetry contract:
- While working, periodically emit fixed progress telemetry as <riido_log>{"code":1001,"args":{}}<end>.
- Use only these active codes: 1001 thinking, 1101 tool collecting, 1102 collection completed count, 1103 tool running, 1104 tool completed.
- For 1101 and 1103 use args label and description. For 1102 use args label, count, and representative_title. For 1104 use args label and summary. If unsure, emit code 1001.
- The label arg must be a short state-neutral noun phrase. Do not include rendered state words such as "수집 중", "조회 완료", "실행 중", "실행", or "완료"; the Riido renderer adds those words.
- Use the tag only for progress telemetry, not for final code blocks.
- Do not invent new codes or free-form user-facing copy.`
}

func TelemetryNativeConfigHardRules() []string {
	return []string{
		`While working, periodically emit progress as <riido_log>{"code":1001,"args":{}}<end>.`,
		"Use only Riido progress message codes 1001, 1101, 1102, 1103, and 1104.",
		"Use args label and description for 1101/1103, label/count/representative_title for 1102, and label/summary for 1104.",
		`Keep label args state-neutral; never include rendered state words like "수집 중", "조회 완료", "실행 중", "실행", or "완료".`,
		"Use <riido_log> only for progress telemetry, not for final code blocks.",
		"Do not invent new Riido progress codes or free-form user-facing copy.",
	}
}
