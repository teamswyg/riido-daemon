package saasplane

import (
	"strings"
)

func normalizedExecutionIDs(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, executionID := range in {
		executionID = strings.TrimSpace(executionID)
		if executionID == "" || seen[executionID] {
			continue
		}
		seen[executionID] = true
		out = append(out, executionID)
	}
	return out
}

func runtimeKindForProvider(provider string) string {
	switch strings.TrimSpace(provider) {
	case "claude":
		return "claude_code"
	default:
		return strings.TrimSpace(provider)
	}
}
