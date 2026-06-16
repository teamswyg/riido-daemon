package agentbridge

import (
	"strings"
)

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
		before, after, ok := strings.Cut(afterStart, telemetryLogEnd)
		if !ok {
			return out
		}
		message := strings.TrimSpace(before)
		if event, ok := progressEventFromTelemetryMessage(message); ok {
			out = append(out, event)
		}
		p.buf = after
	}
}

func suffixThatCanStartTag(s string) string {
	limit := min(len(s), len(telemetryLogStart)-1)
	for n := limit; n > 0; n-- {
		if strings.HasSuffix(s, telemetryLogStart[:n]) {
			return s[len(s)-n:]
		}
	}
	return ""
}
