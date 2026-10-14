package validation

import "strings"

func sanitizeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}
	var builder strings.Builder
	for _, r := range value {
		writeSanitizedIDRune(&builder, r)
	}
	return builder.String()
}

func writeSanitizedIDRune(builder *strings.Builder, r rune) {
	switch {
	case r >= 'a' && r <= 'z':
		builder.WriteRune(r)
	case r >= 'A' && r <= 'Z':
		builder.WriteRune(r)
	case r >= '0' && r <= '9':
		builder.WriteRune(r)
	case r == '-' || r == '_' || r == '.' || r == ':':
		builder.WriteRune(r)
	default:
		builder.WriteRune('-')
	}
}
