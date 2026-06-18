package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func parseUsage(obj map[string]any) agentbridge.Usage {
	intField := func(k string) int {
		switch v := obj[k].(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
		return 0
	}
	return agentbridge.Usage{
		PromptTokens:     intField("input_tokens"),
		CompletionTokens: intField("output_tokens"),
		CacheReadTokens:  intField("cache_read_input_tokens"),
		CacheWriteTokens: intField("cache_creation_input_tokens"),
	}
}
