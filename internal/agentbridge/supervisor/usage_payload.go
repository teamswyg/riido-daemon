package supervisor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func usagePayload(usage agentbridge.Usage) map[string]any {
	return map[string]any{
		"promptTokens":     usage.PromptTokens,
		"completionTokens": usage.CompletionTokens,
		"reasoningTokens":  usage.ReasoningTokens,
		"cacheReadTokens":  usage.CacheReadTokens,
		"cacheWriteTokens": usage.CacheWriteTokens,
	}
}
