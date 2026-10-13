package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func claudeStreamEvents(raw agentbridge.RawEvent) []agentbridge.Event {
	event := mapField(raw.Payload, "event")
	delta := mapField(event, "delta")
	if event["type"] != "content_block_delta" {
		return nil
	}
	switch delta["type"] {
	case "text_delta":
		return []agentbridge.Event{{Kind: agentbridge.EventTextDelta, Text: stringField(delta, "text")}}
	case "thinking_delta":
		return []agentbridge.Event{{Kind: agentbridge.EventThinkingDelta, Text: stringField(delta, "thinking")}}
	}
	return nil
}
