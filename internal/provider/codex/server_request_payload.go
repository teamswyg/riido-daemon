package codex

import (
	"fmt"
	"strconv"
	"strings"
)

func paramsPayload(payload map[string]any) map[string]any {
	p, _ := payload["params"].(map[string]any)
	return p
}

func providerRequestID(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	switch v := payload["id"].(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		if v == float64(int64(v)) {
			return strconv.FormatInt(int64(v), 10)
		}
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
