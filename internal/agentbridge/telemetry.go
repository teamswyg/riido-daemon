package agentbridge

import "strings"

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

// TelemetryParser extracts Riido control-layer telemetry tags from provider
// text deltas. It is provider-neutral and owned by the session actor.
type TelemetryParser struct {
	buf string
}

func NewTelemetryParser() *TelemetryParser {
	return &TelemetryParser{}
}

func ApplyTelemetryContract(provider, prompt, systemPrompt string) (string, string, string) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	switch provider {
	case "claude":
		return strings.TrimSpace(prompt), appendPromptSection(systemPrompt, TelemetryContractInstruction()), TelemetryPlacementSystemPrompt
	case "openclaw":
		return strings.TrimSpace(prompt), appendPromptSection(systemPrompt, TelemetryContractInstruction()), TelemetryPlacementSystemPromptInline
	default:
		return InjectTelemetryContract(prompt), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt
	}
}

func ApplyRuntimeInstructionContract(provider, prompt, systemPrompt, agentInstruction string) (string, string, string, string) {
	provider = strings.TrimSpace(strings.ToLower(provider))
	agentSection := AgentInstructionContract(agentInstruction)
	switch provider {
	case "claude":
		system := appendPromptSection(systemPrompt, agentSection)
		system = appendPromptSection(system, TelemetryContractInstruction())
		return strings.TrimSpace(prompt), system, TelemetryPlacementSystemPrompt, placementForSection(agentSection, TelemetryPlacementSystemPrompt)
	case "openclaw":
		system := appendPromptSection(systemPrompt, agentSection)
		system = appendPromptSection(system, TelemetryContractInstruction())
		return strings.TrimSpace(prompt), system, TelemetryPlacementSystemPromptInline, placementForSection(agentSection, TelemetryPlacementSystemPromptInline)
	default:
		return InjectPromptSections(prompt, agentSection, TelemetryContractInstruction()), strings.TrimSpace(systemPrompt), TelemetryPlacementPrompt, placementForSection(agentSection, TelemetryPlacementPrompt)
	}
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
- While working, periodically emit progress as <riido_log>short Korean status<end>.
- Use the tag only for progress telemetry, not for final code blocks.
- Keep each telemetry message under 120 characters.`
}

func TelemetryNativeConfigHardRules() []string {
	return []string{
		"While working, periodically emit progress as <riido_log>short Korean status<end>.",
		"Use <riido_log> only for progress telemetry, not for final code blocks.",
		"Keep each Riido telemetry message under 120 characters.",
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
		if message != "" {
			out = append(out, Event{Kind: EventProgress, Text: message})
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
