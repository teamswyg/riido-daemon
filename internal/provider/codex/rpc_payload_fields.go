package codex

// rpcID extracts the JSON-RPC "id" field as int64. JSON numbers come back as
// float64; tests and helpers may build payloads directly with int or int64.
func rpcID(p map[string]any) (int64, bool) {
	switch v := p["id"].(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	}
	return 0, false
}

func mapField(p map[string]any, key string) map[string]any {
	if p == nil {
		return nil
	}
	m, _ := p[key].(map[string]any)
	return m
}

func threadIDFromResult(result map[string]any) string {
	if id := stringField(result, "thread_id"); id != "" {
		return id
	}
	thread := mapField(result, "thread")
	if id := stringField(thread, "id"); id != "" {
		return id
	}
	return stringField(thread, "sessionId")
}
