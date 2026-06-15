package toolargs

import (
	"strings"
)

func joinKey(prefix, key string) string {
	key = strings.TrimSpace(key)
	if key == "" {
		return prefix
	}
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func truncate(value string) string {
	runes := []rune(value)
	if len(runes) <= maxValueRunes {
		return value
	}
	return string(runes[:maxValueRunes])
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, ".", "_")
	return value
}

func nilIfEmpty(args map[string]string) map[string]string {
	if len(args) == 0 {
		return nil
	}
	return args
}
