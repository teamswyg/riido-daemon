package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (d *protocolDriver) failedEvents(message string) []agentbridge.Event {
	if message == "" {
		message = "codex runtime error"
	}
	return []agentbridge.Event{
		{Kind: agentbridge.EventError, Err: message},
		{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: agentbridge.ResultFailed,
				Error:  message,
			},
		},
	}
}
