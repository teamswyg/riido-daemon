package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func codexTurnCompletedEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventResult,
		Result: agentbridge.Result{
			Status: agentbridge.ResultCompleted,
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
