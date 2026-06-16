package claude

import (
	"encoding/json"
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func translateResult(raw agentbridge.RawEvent) []agentbridge.Event {
	var out []agentbridge.Event
	if usage, ok := raw.Payload["usage"].(map[string]any); ok {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(usage)})
	}
	subtype := wireResultSubtype(stringField(raw.Payload, "subtype"))
	status := agentbridge.ResultCompleted
	switch subtype {
	case wireResultSubtypeError, wireResultSubtypeExecutionError:
		status = agentbridge.ResultFailed
	case wireResultSubtypeCancelled:
		status = agentbridge.ResultCancelled
	case wireResultSubtypeMaxTurns:
		status = agentbridge.ResultBlocked
	}
	result := agentbridge.Result{
		Status: status,
		Output: stringField(raw.Payload, "result"),
		Error:  stringField(raw.Payload, "error"),
	}
	out = append(out, agentbridge.Event{Kind: agentbridge.EventResult, Result: result})
	return out
}

func firstToolInput(request map[string]any) any {
	for _, key := range []string{"tool_input", "input", "args"} {
		if value, ok := request[key]; ok {
			return value
		}
	}
	return nil
}

func parseUsage(obj map[string]any) agentbridge.Usage {
	intField := func(k string) int {
		switch v := obj[k].(type) {
		case float64:
			return int(v)
		case int:
			return v
		}
		return 0
	}
	return agentbridge.Usage{
		PromptTokens:     intField("input_tokens"),
		CompletionTokens: intField("output_tokens"),
		CacheReadTokens:  intField("cache_read_input_tokens"),
		CacheWriteTokens: intField("cache_creation_input_tokens"),
	}
}

// BuildProviderInput serializes reducer approval commands into Claude's
// stream-json control_response shape. The response schema follows the Claude
// Agent SDK control router: control_response.response carries the request_id
// and a behavior-specific payload.
func BuildProviderInput(cmd agentbridge.Command) ([]byte, error) {
	requestID := cmd.ProviderRequestID
	if requestID == "" {
		return nil, fmt.Errorf("claude: provider request id is required for %s", cmd.Kind)
	}
	var response map[string]any
	switch cmd.Kind {
	case agentbridge.CommandApproveTool:
		response = map[string]any{
			"behavior":     "allow",
			"updatedInput": map[string]any{},
		}
	case agentbridge.CommandRejectTool:
		reason := cmd.Reason
		if reason == "" {
			reason = "Permission denied"
		}
		response = map[string]any{
			"behavior": "deny",
			"message":  reason,
		}
	default:
		return nil, fmt.Errorf("claude: unsupported provider input command %s", cmd.Kind)
	}
	frame := map[string]any{
		"type": "control_response",
		"response": map[string]any{
			"subtype":    "success",
			"request_id": requestID,
			"response":   response,
		},
	}
	body, err := json.Marshal(frame)
	if err != nil {
		return nil, fmt.Errorf("claude: marshal control_response: %w", err)
	}
	return append(body, '\n'), nil
}

func stringField(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	s, _ := m[key].(string)
	return s
}

// claudeRateLimitDetail extracts a human-readable detail from a Claude Code
// rate_limit_event payload, tolerating both a flat shape and a nested
// "rate_limit" object. Falls back to a generic note so the Warning is never
// empty.
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

func mapField(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	nested, _ := m[key].(map[string]any)
	return nested
}
