package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func parseUsage(m map[string]any) agentbridge.Usage {
	intField := func(keys ...string) int {
		for _, k := range keys {
			if v := intValue(m[k]); v != 0 {
				return v
			}
		}
		return 0
	}
	return agentbridge.Usage{
		PromptTokens:     intField("prompt_tokens", "input"),
		CompletionTokens: intField("completion_tokens", "output"),
		ReasoningTokens:  intField("reasoning_tokens", "reasoning"),
		CacheReadTokens:  intField("cache_read_tokens", "cacheRead"),
		CacheWriteTokens: intField("cache_write_tokens", "cacheWrite"),
	}
}
