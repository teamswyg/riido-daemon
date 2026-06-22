package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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
	output := stringField(raw.Payload, "result")
	errText := stringField(raw.Payload, "error")
	if isError && errText == "" {
		errText = output
	}
	return agentbridge.Result{
		Status: claudeResultStatus(wireResultSubtype(stringField(raw.Payload, "subtype")), isError),
		Output: output,
		Error:  errText,
	}
}

func claudeResultStatus(subtype wireResultSubtype, isError bool) agentbridge.ResultStatus {
	if isError {
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
