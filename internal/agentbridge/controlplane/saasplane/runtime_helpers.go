package saasplane

import (
	"strings"

	providercatalog "github.com/teamswyg/riido-contracts/provider/catalog"
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
	switch providercatalog.Normalize(provider) {
	case providercatalog.KindClaude:
		return string(providercatalog.KindClaudeCode)
	default:
		return providercatalog.String(provider)
	}
}
