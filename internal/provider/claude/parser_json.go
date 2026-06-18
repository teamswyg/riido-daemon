package claude

import (
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func decodeJSONRawEvent(ev agentbridge.RawEvent, trimmed []byte) agentbridge.RawEvent {
	if trimmed[0] != '{' && trimmed[0] != '[' {
		ev.Type = "malformed"
		return ev
	}

	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		ev.Type = "malformed"
		return ev
	}
	ev.Payload = payload
	ev.Type = rawEventType(payload)
	return ev
}

func rawEventType(payload map[string]any) string {
	if t, ok := payload["type"].(string); ok {
		return t
	}
	return "unknown"
}
