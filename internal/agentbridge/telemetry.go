package agentbridge

import (
	"errors"
	"strings"
)

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
	switch strategy.AgentInstructionPlacement {
	case TelemetryPlacementSystemPrompt, TelemetryPlacementSystemPromptInline:
		system := appendPromptSection(systemPrompt, agentSection)
		system = appendPromptSection(system, TelemetryContractInstruction())
		return strings.TrimSpace(prompt), system, strategy.TelemetryPlacement, placementForSection(agentSection, strategy.AgentInstructionPlacement)
	case TelemetryPlacementPrompt:
		return InjectPromptSections(prompt, agentSection, TelemetryContractInstruction()), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt, placementForSection(agentSection, TelemetryPlacementPrompt)
	default:
		return InjectPromptSections(prompt, agentSection, TelemetryContractInstruction()), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt, placementForSection(agentSection, TelemetryPlacementPrompt)
	}
}

func RuntimeInstructionStrategies() []RuntimeInstructionStrategy {
	strategies := []RuntimeInstructionStrategy{
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
	out := make([]RuntimeInstructionStrategy, len(strategies))
	copy(out, strategies)
	return out
}

func RuntimeInstructionStrategyForProvider(provider string) RuntimeInstructionStrategy {
	provider = normalizeProviderName(provider)
	for _, strategy := range RuntimeInstructionStrategies() {
		if strategy.Provider == provider {
			return strategy
		}
	}
	return RuntimeInstructionStrategy{
		Provider:                  provider,
		AgentInstructionPlacement: TelemetryPlacementPrompt,
		TelemetryPlacement:        TelemetryPlacementPrompt,
		EffectivenessGate:         "opt-in-real-provider-probe",
	}
}

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

func InjectTelemetryContract(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	if strings.Contains(prompt, telemetryLogStart) && strings.Contains(prompt, telemetryLogEnd) {
		return prompt
	}
	return strings.TrimSpace(TelemetryContractInstruction() + "\n\nUser task:\n" + prompt)
}

func InjectPromptSections(prompt string, sections ...string) string {
	prompt = strings.TrimSpace(prompt)
	var out []string
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" || strings.Contains(prompt, section) {
			continue
		}
		out = append(out, section)
	}
	if prompt != "" {
		out = append(out, "User task:\n"+prompt)
	}
	return strings.TrimSpace(strings.Join(out, "\n\n"))
}

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

func appendPromptSection(existing, section string) string {
	existing = strings.TrimSpace(existing)
	section = strings.TrimSpace(section)
	if existing == "" {
		return section
	}
	if section == "" || strings.Contains(existing, section) {
		return existing
	}
	return existing + "\n\n" + section
}

func placementForSection(section, placement string) string {
	if strings.TrimSpace(section) == "" {
		return ""
	}
	return placement
}

func normalizeProviderName(provider string) string {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == "" {
		return "default"
	}
	return provider
}

func (p *TelemetryParser) Feed(text string) []Event {
	if p == nil || text == "" {
		return nil
	}
	p.buf += text
	if len(p.buf) > 64*1024 {
		p.buf = p.buf[len(p.buf)-64*1024:]
	}
	out := []Event{}
	for {
		start := strings.Index(p.buf, telemetryLogStart)
		if start < 0 {
			p.buf = suffixThatCanStartTag(p.buf)
			return out
		}
		if start > 0 {
			p.buf = p.buf[start:]
		}
		afterStart := p.buf[len(telemetryLogStart):]
		end := strings.Index(afterStart, telemetryLogEnd)
		if end < 0 {
			return out
		}
		message := strings.TrimSpace(afterStart[:end])
		if event, ok := progressEventFromTelemetryMessage(message); ok {
			out = append(out, event)
		}
		p.buf = afterStart[end+len(telemetryLogEnd):]
	}
}

func suffixThatCanStartTag(s string) string {
	max := len(telemetryLogStart) - 1
	if len(s) < max {
		max = len(s)
	}
	for n := max; n > 0; n-- {
		if strings.HasSuffix(s, telemetryLogStart[:n]) {
			return s[len(s)-n:]
		}
	}
	return ""
}
