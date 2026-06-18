package openclaw

func intValue(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func mapField(m map[string]any, key string) (map[string]any, bool) {
	if m == nil {
		return nil, false
	}
	child, ok := m[key].(map[string]any)
	return child, ok
}

func stringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}
