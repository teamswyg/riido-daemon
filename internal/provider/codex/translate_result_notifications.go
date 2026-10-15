package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func codexTurnCompletedEvent(p map[string]any) []agentbridge.Event {
	turn := mapField(p, "turn")
	status := agentbridge.ResultCompleted
	switch stringField(turn, "status") {
	case "failed":
		status = agentbridge.ResultFailed
	case "interrupted":
		status = agentbridge.ResultCancelled
	case "", "completed":
	default:
		status = agentbridge.ResultFailed
	}
	return []agentbridge.Event{{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: status,
			Error:  stringField(mapField(turn, "error"), "message"),
			Output: stringField(p, "output"),
		},
	}}
}

func codexTurnFailedEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: agentbridge.ResultFailed,
			Error:  stringField(p, "message"),
		},
	}}
}
