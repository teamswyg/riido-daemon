package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func codexUsageDeltaEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{
		PromptTokens:     intField(p, "input_tokens"),
		CompletionTokens: intField(p, "output_tokens"),
		ReasoningTokens:  intField(p, "reasoning_tokens"),
	}}}
}

func codexThreadTokenUsageEvent(p map[string]any) []agentbridge.Event {
	total := mapField(mapField(p, "tokenUsage"), "total")
	return []agentbridge.Event{{Kind: agentbridge.EventUsageDelta, Usage: agentbridge.Usage{
		PromptTokens:     intField(total, "inputTokens"),
		CompletionTokens: intField(total, "outputTokens"),
		ReasoningTokens:  intField(total, "reasoningOutputTokens"),
		CacheReadTokens:  intField(total, "cachedInputTokens"),
	}}}
}
