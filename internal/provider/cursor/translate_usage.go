package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func parseUsage(m map[string]any) agentbridge.Usage {
	return agentbridge.Usage{
		PromptTokens:     intField(m, "input_tokens"),
		CompletionTokens: intField(m, "output_tokens"),
	}
}

func intField(m map[string]any, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}
