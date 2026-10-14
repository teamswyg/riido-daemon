package textutil

import "strings"

// Default returns fallback when value is empty after trimming whitespace.
// Otherwise it returns the trimmed value.
func Default(value, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return fallback
}

// FirstNonEmpty returns the first argument that is not empty after trimming
// whitespace. The returned value is the original argument, preserving caller
// formatting.
func FirstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// FirstNonEmptyTrimmed returns the trimmed form of the first argument that is
// not empty after trimming whitespace.
func FirstNonEmptyTrimmed(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
