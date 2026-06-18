package agentbridge

import "strings"

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
