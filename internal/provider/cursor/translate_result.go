package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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
			Error:  stringField(p, "error"),
		},
	})
}

func resultStatus(p map[string]any) agentbridge.ResultStatus {
	switch wireResultSubtype(stringField(p, "subtype")) {
	case wireResultSubtypeError, wireResultSubtypeExecutionError:
		return agentbridge.ResultFailed
	case wireResultSubtypeCancelled:
		return agentbridge.ResultCancelled
	default:
		return agentbridge.ResultCompleted
	}
}
