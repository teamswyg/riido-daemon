package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func codexAgentTextDeltaEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventTextDelta,
		Text: stringField(p, "text"),
	}}
}

func codexItemTextDeltaEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventTextDelta,
		Text: stringField(p, "delta"),
	}}
}

func codexReasoningDeltaEvent(p map[string]any) []agentbridge.Event {
	return []agentbridge.Event{{
		Kind: agentbridge.EventThinkingDelta,
		Text: stringField(p, "text"),
	}}
}
