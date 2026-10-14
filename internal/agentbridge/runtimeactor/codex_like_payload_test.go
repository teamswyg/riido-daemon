package runtimeactor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func jsonRPCFrameType(payload map[string]any) string {
	method, _ := payload["method"].(string)
	_, hasID := payload["id"]
	switch {
	case method != "" && hasID:
		return "server_request:" + method
	case method != "":
		return "notification:" + method
	default:
		return "response"
	}
}

func payloadInt64(payload map[string]any, key string) (int64, bool) {
	switch v := payload[key].(type) {
	case float64:
		return int64(v), true
	case int:
		return int64(v), true
	case int64:
		return v, true
	default:
		return 0, false
	}
}

func completedProtocolResult(raw agentbridge.RawEvent) agentbridge.Event {
	return agentbridge.Event{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: agentbridge.ResultCompleted,
			Output: stringFromPayload(raw.Payload, "output"),
		},
	}
}

func stringFromPayload(payload map[string]any, key string) string {
	if params, ok := payload["params"].(map[string]any); ok {
		if value, ok := params[key].(string); ok {
			return value
		}
	}
	value, _ := payload[key].(string)
	return value
}
