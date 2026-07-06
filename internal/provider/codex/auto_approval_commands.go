package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func autoApproveServerRequestEvents(method codexMethod, events []agentbridge.Event) []agentbridge.Event {
	var out []agentbridge.Event
	for _, event := range events {
		if event.Kind != agentbridge.EventToolApprovalNeeded {
			out = append(out, event)
			continue
		}
		text := "codex auto-approved server_request: " + string(method)
		if event.Tool.ProviderRequestID == "" {
			text = "codex approval server_request missing id: " + string(method)
		}
		out = append(out, agentbridge.Event{Kind: agentbridge.EventLog, Text: text})
	}
	return out
}

func autoApproveServerRequestCommands(events []agentbridge.Event) []agentbridge.Command {
	var cmds []agentbridge.Command
	for _, event := range events {
		if event.Kind != agentbridge.EventToolApprovalNeeded {
			continue
		}
		if event.Tool.ProviderRequestID == "" {
			continue
		}
		cmds = append(cmds, agentbridge.Command{
			Kind:              agentbridge.CommandApproveTool,
			ToolID:            event.Tool.ID,
			ProviderRequestID: event.Tool.ProviderRequestID,
		})
	}
	return cmds
}
