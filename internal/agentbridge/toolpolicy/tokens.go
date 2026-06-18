package toolpolicy

import "strings"

func matchesAny(kind, name string, candidates ...string) bool {
	for _, candidate := range candidates {
		normalized := normalizeToolToken(candidate)
		if kind == normalized || name == normalized {
			return true
		}
	}
	return false
}

func normalizeToolToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}
