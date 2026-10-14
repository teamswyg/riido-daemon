package claude

import (
	"strings"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func translateResult(raw agentbridge.RawEvent) []agentbridge.Event {
	var out []agentbridge.Event
	if usage, ok := raw.Payload["usage"].(map[string]any); ok {
		out = append(out, agentbridge.Event{Kind: agentbridge.EventUsageDelta, Usage: parseUsage(usage)})
	}
	out = append(out, agentbridge.Event{Kind: agentbridge.EventResult, Result: claudeResult(raw)})
	return out
}

func claudeResult(raw agentbridge.RawEvent) agentbridge.Result {
	isError := boolField(raw.Payload, "is_error")
	output := normalizeClaudeErrorMessage(stringField(raw.Payload, "result"))
	errText := stringField(raw.Payload, "error")
	status := claudeResultStatus(wireResultSubtype(stringField(raw.Payload, "subtype")), isError)
	if status != agentbridge.ResultCompleted && errText == "" {
		if messages, ok := raw.Payload["errors"].([]any); ok {
			var parts []string
			for _, value := range messages {
				if message, ok := value.(string); ok && message != "" {
					parts = append(parts, message)
				}
			}
			errText = strings.Join(parts, "\n")
		}
	}
	if status != agentbridge.ResultCompleted && errText == "" {
		errText = output
	}
	if status != agentbridge.ResultCompleted && errText == "" {
		errText = "claude result: " + stringField(raw.Payload, "subtype")
	}
	// Preserve the native result detail; runtime summaries are separately bounded.
	return agentbridge.Result{
		Status: status,
		Output: output,
		Error:  errText,
	}
}

func claudeResultStatus(subtype wireResultSubtype, isError bool) agentbridge.ResultStatus {
	if subtype == "error_max_turns" || subtype == "error_max_budget_usd" {
		return agentbridge.ResultBlocked
	}
	if isError || strings.HasPrefix(string(subtype), "error") {
		return agentbridge.ResultFailed
	}
	switch subtype {
	case wireResultSubtypeError, wireResultSubtypeExecutionError:
		return agentbridge.ResultFailed
	case wireResultSubtypeCancelled:
		return agentbridge.ResultCancelled
	case wireResultSubtypeMaxTurns:
		return agentbridge.ResultBlocked
	default:
		return agentbridge.ResultCompleted
	}
}
