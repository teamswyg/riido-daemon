package cursor

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func toolStartedFromPayload(payload map[string]any) agentbridge.Event {
	name := stringField(payload, "name")
	return agentbridge.Event{
		Kind: agentbridge.EventToolCallStarted,
		Tool: agentbridge.ToolRef{
			ID:   stringField(payload, "id"),
			Name: name,
			Kind: name,
			Args: toolargs.FromValue(firstToolInput(payload)),
		},
	}
}

func toolResultFromPayload(payload map[string]any) agentbridge.Event {
	kind := agentbridge.EventToolCallCompleted
	if isErr, _ := payload["is_error"].(bool); isErr {
		kind = agentbridge.EventToolCallFailed
	}
	return agentbridge.Event{Kind: kind, Tool: agentbridge.ToolRef{ID: stringField(payload, "tool_use_id")}}
}

func firstToolInput(payload map[string]any) any {
	for _, key := range []string{"input", "tool_input", "args"} {
		if value, ok := payload[key]; ok {
			return value
		}
	}
	return nil
}
