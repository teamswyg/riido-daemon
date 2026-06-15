package codex

func stringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

func intField(m map[string]any, key string) int {
	if m == nil {
		return 0
	}
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

func errMessage(payload map[string]any) string {
	e, ok := payload["error"].(map[string]any)
	if !ok {
		return ""
	}
	return stringField(e, "message")
}
