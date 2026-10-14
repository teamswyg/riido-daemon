package claude

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func translateAssistantMessage(raw agentbridge.RawEvent) []agentbridge.Event {
	var out []agentbridge.Event
	for _, obj := range claudeMessageContent(raw) {
		switch wireContentType(stringField(obj, "type")) {
		case wireContentText:
			out = append(out, agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: stringField(obj, "text")})
		case wireContentThinking:
			out = append(out, agentbridge.Event{Kind: agentbridge.EventThinkingDelta, Text: stringField(obj, "thinking")})
		case wireContentToolUse:
			out = append(out, claudeToolUseStartedEvent(obj))
		default:
			continue
		}
	}
	return out
}

func claudeToolUseStartedEvent(obj map[string]any) agentbridge.Event {
	name := stringField(obj, "name")
	return agentbridge.Event{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   stringField(obj, "id"),
			Name: name,
			Kind: name,
			Args: toolargs.FromValue(obj["input"]),
		},
	}
}
