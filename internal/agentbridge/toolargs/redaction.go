package toolargs

import "strings"

// IsSensitiveKey reports whether key names material that must be redacted.
func IsSensitiveKey(key string) bool {
	normalized := normalizeKey(key)
	for _, token := range sensitiveKeyTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

// HasRedactedValue reports whether args contains a redacted value marker.
func HasRedactedValue(args map[string]string) bool {
	for _, value := range args {
		if IsRedactedValue(value) {
			return true
		}
	}
	return false
}

// IsRedactedValue reports whether value is the ToolRef.Args redaction marker.
func IsRedactedValue(value string) bool {
	return strings.TrimSpace(value) == RedactedValue
}
