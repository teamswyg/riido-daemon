package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (d *protocolDriver) failedEvents(message string) []agentbridge.Event {
	return codexFailureEvents(nil, message)
}

func codexFailureEvents(details map[string]any, message string) []agentbridge.Event {
	if message == "" {
		message = "codex runtime error"
	}
	return []agentbridge.Event{
		{Kind: agentbridge.EventError, Err: message},
		{
			Kind: agentbridge.EventResult,
			Result: agentbridge.Result{
				Status: codexFailureStatus(details, message),
				Error:  message,
			},
		},
	}
}
