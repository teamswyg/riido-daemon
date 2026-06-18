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
	return agentbridge.Result{
		Status: claudeResultStatus(wireResultSubtype(stringField(raw.Payload, "subtype"))),
		Output: stringField(raw.Payload, "result"),
		Error:  stringField(raw.Payload, "error"),
	}
}

func claudeResultStatus(subtype wireResultSubtype) agentbridge.ResultStatus {
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
