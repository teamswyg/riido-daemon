package openclaw

import (
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func closeFullResultStdout(trimmed []byte) []agentbridge.RawEvent {
	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return []agentbridge.RawEvent{
			rawBytesEvent(agentbridge.RawSourceClose, "malformed", trimmed),
		}
	}
	return []agentbridge.RawEvent{
		rawPayloadEvent(agentbridge.RawSourceClose, "full_result", payload, trimmed),
	}
}
