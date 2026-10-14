package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func translateUserMessage(raw agentbridge.RawEvent) []agentbridge.Event {
	var out []agentbridge.Event
	for _, obj := range claudeMessageContent(raw) {
		if wireContentType(stringField(obj, "type")) == wireContentToolResult {
			out = append(out, claudeToolResultEvent(obj))
		}
	}
	return out
}

func claudeToolResultEvent(obj map[string]any) agentbridge.Event {
	isErr, _ := obj["is_error"].(bool)
	kind := agentbridge.EventToolCallCompleted
	if isErr {
		kind = agentbridge.EventToolCallFailed
	}
	return agentbridge.Event{
		Kind: kind,
		Tool: agentbridge.ToolRef{ID: stringField(obj, "tool_use_id")},
	}
}
