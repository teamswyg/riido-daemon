package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func translateAssistant(p map[string]any) []agentbridge.Event {
	content, _ := p["content"].([]any)
	var out []agentbridge.Event
	for _, item := range content {
		if ev, ok := assistantItemEvent(item); ok {
			out = append(out, ev)
		}
	}
	return out
}

func assistantItemEvent(item any) (agentbridge.Event, bool) {
	obj, ok := item.(map[string]any)
	if !ok {
		return agentbridge.Event{}, false
	}
	switch wireContentType(stringField(obj, "type")) {
	case wireContentText, wireContentOutputText:
		return agentbridge.Event{Kind: agentbridge.EventTextDelta, Text: stringField(obj, "text")}, true
	case wireContentThinking:
		return agentbridge.Event{Kind: agentbridge.EventThinkingDelta, Text: stringField(obj, "text")}, true
	case wireContentToolUse:
		return toolStartedFromPayload(obj), true
	default:
		return agentbridge.Event{}, false
	}
}
