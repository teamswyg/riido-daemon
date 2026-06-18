package claude

func claudeRateLimitDetail(payload map[string]any) string {
	for _, scope := range []map[string]any{payload, mapField(payload, "rate_limit")} {
		if scope == nil {
			continue
		}
		for _, key := range []string{"message", "status", "resets_at", "resetsAt", "retry_after", "retryAfter"} {
			if v := stringField(scope, key); v != "" {
				return v
			}
		}
	}
	return "upstream rate limit reached"
}
