package codex

// classifyJSONRPC tags a JSON-RPC 2.0 frame:
//   - "notification:<method>" — method present, no id
//   - "server_request:<method>" — method present, id present
//   - "response" — result present, no method
//   - "error" — error present, no method
//   - "unknown" — none of the above
func classifyJSONRPC(payload map[string]any) string {
	method, hasMethod := payload["method"].(string)
	_, hasID := payload["id"]
	if hasMethod {
		if hasID {
			return "server_request:" + method
		}
		return "notification:" + method
	}
	if _, hasResult := payload["result"]; hasResult {
		return "response"
	}
	if _, hasError := payload["error"]; hasError {
		return "error"
	}
	return "unknown"
}
