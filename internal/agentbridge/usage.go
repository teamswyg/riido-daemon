package agentbridge

// Usage is the provider-neutral token-usage accumulator. Each provider
// reports usage in its own schema; adapters normalize into this struct
// (docs/20-domain/provider-runtime.md §5.5).
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	ReasoningTokens  int
	CacheReadTokens  int
	CacheWriteTokens int
}

// Add returns the field-wise sum of u and other.
func (u Usage) Add(other Usage) Usage {
	return Usage{
		PromptTokens:     u.PromptTokens + other.PromptTokens,
		CompletionTokens: u.CompletionTokens + other.CompletionTokens,
		ReasoningTokens:  u.ReasoningTokens + other.ReasoningTokens,
		CacheReadTokens:  u.CacheReadTokens + other.CacheReadTokens,
		CacheWriteTokens: u.CacheWriteTokens + other.CacheWriteTokens,
	}
}
