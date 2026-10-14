package agentbridge

import (
	"errors"
	"strings"
)

func BuildInstructionEffectivenessProbe(provider string) InstructionEffectivenessProbe {
	strategy := RuntimeInstructionStrategyForProvider(provider)
	marker := "RIIDO_INSTRUCTION_ACK:" + strategy.Provider
	agentInstruction := "검증용 지시입니다. 응답 첫 줄에 `" + marker + "`를 그대로 포함한 뒤 사용자 요청을 수행하세요."
	userPrompt := "Riido agent instruction 전달 경로 검증입니다. 한 문장으로 지시 수신 여부를 한국어로 답하세요."
	prompt, systemPrompt, telemetryPlacement, instructionPlacement := ApplyRuntimeInstructionContract(strategy.Provider, userPrompt, "", agentInstruction)
	return InstructionEffectivenessProbe{
		Provider:                  strategy.Provider,
		Prompt:                    prompt,
		SystemPrompt:              systemPrompt,
		ExpectedMarker:            marker,
		AgentInstructionPlacement: instructionPlacement,
		TelemetryPlacement:        telemetryPlacement,
	}
}

func ValidateInstructionEffectivenessOutput(probe InstructionEffectivenessProbe, output string) error {
	if strings.TrimSpace(probe.ExpectedMarker) == "" {
		return errors.New("instruction effectiveness probe marker is empty")
	}
	if !strings.Contains(output, probe.ExpectedMarker) {
		return errors.New("instruction effectiveness marker missing from provider output")
	}
	return nil
}
