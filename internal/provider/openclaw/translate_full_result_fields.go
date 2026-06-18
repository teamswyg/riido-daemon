package openclaw

func fullResultSessionID(p map[string]any) string {
	if sid := stringField(p, "session_id"); sid != "" {
		return sid
	}
	if meta, ok := mapField(p, "meta"); ok {
		if agentMeta, ok := mapField(meta, "agentMeta"); ok {
			return stringField(agentMeta, "sessionId")
		}
	}
	return ""
}

func fullResultUsage(p map[string]any) (map[string]any, bool) {
	if usage, ok := mapField(p, "usage"); ok {
		return usage, true
	}
	if meta, ok := mapField(p, "meta"); ok {
		if usage, ok := fullResultAgentMetaUsage(meta); ok {
			return usage, true
		}
	}
	return nil, false
}

func fullResultAgentMetaUsage(meta map[string]any) (map[string]any, bool) {
	agentMeta, ok := mapField(meta, "agentMeta")
	if !ok {
		return nil, false
	}
	if usage, ok := mapField(agentMeta, "usage"); ok {
		return usage, true
	}
	if usage, ok := mapField(agentMeta, "lastCallUsage"); ok {
		return usage, true
	}
	return nil, false
}

func fullResultText(p map[string]any) string {
	if text := stringField(p, "text"); text != "" {
		return text
	}
	payloads, _ := p["payloads"].([]any)
	for _, payload := range payloads {
		m, ok := payload.(map[string]any)
		if !ok {
			continue
		}
		if text := stringField(m, "text"); text != "" {
			return text
		}
	}
	return ""
}
