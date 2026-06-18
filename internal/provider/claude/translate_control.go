package claude

import (
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/toolargs"
)

func translateControlRequest(raw agentbridge.RawEvent) []agentbridge.Event {
	request, _ := raw.Payload["request"].(map[string]any)
	if wireControlSubtype(stringField(request, "subtype")) == wireControlPermissionRequest {
		return []agentbridge.Event{claudeToolApprovalNeededEvent(raw, request)}
	}
	return []agentbridge.Event{{
		Kind: agentbridge.EventLog,
		Text: "claude unknown control_request subtype: " + stringField(request, "subtype"),
	}}
}

func claudeToolApprovalNeededEvent(raw agentbridge.RawEvent, request map[string]any) agentbridge.Event {
	name := stringField(request, "tool_name")
	return agentbridge.Event{
		Kind: agentbridge.EventToolApprovalNeeded,
		Tool: agentbridge.ToolRef{
			ID:                stringField(request, "tool_use_id"),
			Name:              name,
			Kind:              name,
			Args:              toolargs.FromValue(firstToolInput(request)),
			ProviderRequestID: stringField(raw.Payload, "request_id"),
		},
	}
}

func firstToolInput(request map[string]any) any {
	for _, key := range []string{"tool_input", "input", "args"} {
		if value, ok := request[key]; ok {
			return value
		}
	}
	return nil
}
