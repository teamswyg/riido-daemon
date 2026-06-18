package openclaw

import "strings"

func sanitizeReason(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return "openclaw --version failed"
	}

	s := normalizeReasonWhitespace(raw)
	const maxLen = 300
	if len(s) > maxLen {
		s = s[:maxLen] + "..."
	}
	return s
}

func normalizeReasonWhitespace(raw string) string {
	s := strings.ReplaceAll(raw, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}
