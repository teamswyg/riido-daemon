package main

import "strings"

func localBackendUnavailable(text string) bool {
	return strings.Contains(text, "local model backend unavailable") ||
		strings.Contains(text, "connection refused by the provider endpoint") ||
		strings.Contains(text, "failovererror") ||
		strings.Contains(text, "provider ollama") ||
		strings.Contains(text, "all models failed") ||
		strings.Contains(text, "model-fallback") ||
		strings.Contains(text, "cooldown")
}
