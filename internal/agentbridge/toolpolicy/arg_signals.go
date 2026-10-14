package toolpolicy

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func hasSensitiveArgSignal(args map[string]string) bool {
	for key := range args {
		if toolargs.IsSensitiveKey(key) {
			return true
		}
	}
	return toolargs.HasRedactedValue(args)
}

func argsContainNetworkEgress(args map[string]string) bool {
	for key, value := range args {
		if argValueHasNetworkSignal(key, value) {
			return true
		}
	}
	return false
}

func argValueHasNetworkSignal(key, value string) bool {
	normalizedKey := normalizeToolToken(key)
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	if strings.Contains(normalizedValue, "https://") || strings.Contains(normalizedValue, "http://") {
		return true
	}
	if strings.Contains(normalizedKey, "url") || strings.Contains(normalizedKey, "uri") || strings.Contains(normalizedKey, "endpoint") {
		return strings.TrimSpace(value) != ""
	}
	return commandContainsNetworkEgress(value)
}

func argsTouchProtectedPath(args map[string]string) bool {
	for key, value := range args {
		if pathLikeArgKey(key) && isProtectedPath(value) {
			return true
		}
	}
	return false
}

func pathLikeArgKey(key string) bool {
	normalized := normalizeToolToken(key)
	return strings.Contains(normalized, "path") ||
		strings.Contains(normalized, "file") ||
		strings.Contains(normalized, "target")
}
