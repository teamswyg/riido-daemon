package openclaw

import "strings"

// buildMessage inlines the system prompt above the user prompt when both
// are present, separated by a blank line. When system prompt is empty,
// the user prompt is returned verbatim.
func buildMessage(systemPrompt, userPrompt string) string {
	system := strings.TrimSpace(systemPrompt)
	user := strings.TrimSpace(userPrompt)
	if system == "" {
		return userPrompt
	}
	return system + "\n\n" + user
}
