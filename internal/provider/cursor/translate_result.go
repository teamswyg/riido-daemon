package cursor

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func translateResult(p map[string]any) []agentbridge.Event {
	var out []agentbridge.Event
	if usage, ok := p["usage"].(map[string]any); ok {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(usage)})
	}
	return append(out, agentbridge.Event{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: resultStatus(p),
			Output: stringField(p, "result"),
			Error:  resultError(p),
		},
	})
}

func resultStatus(p map[string]any) agentbridge.ResultStatus {
	if stringField(p, "subtype") == string(wireResultSubtypeCancelled) {
		return agentbridge.ResultCancelled
	}
	if failed, _ := p["is_error"].(bool); failed || resultError(p) != "" || strings.HasPrefix(stringField(p, "subtype"), "error_") {
		return agentbridge.ResultFailed
	}
	switch wireResultSubtype(stringField(p, "subtype")) {
	case wireResultSubtypeError, wireResultSubtypeExecutionError:
		return agentbridge.ResultFailed
	case wireResultSubtypeCancelled:
		return agentbridge.ResultCancelled
	default:
		return agentbridge.ResultCompleted
	}
}

func resultError(p map[string]any) string {
	if detail := stringField(p, "error"); detail != "" {
		return detail
	}
	var details []string
	if list, ok := p["errors"].([]any); ok {
		for _, item := range list {
			if text, ok := item.(string); ok && text != "" {
				details = append(details, text)
			}
		}
	}
	if len(details) > 0 {
		return strings.Join(details, "; ")
	}
	if failed, _ := p["is_error"].(bool); failed || strings.HasPrefix(stringField(p, "subtype"), "error") {
		if output := stringField(p, "result"); output != "" {
			return output
		}
		return "Cursor execution failed: " + stringField(p, "subtype")
	}
	return ""
}
