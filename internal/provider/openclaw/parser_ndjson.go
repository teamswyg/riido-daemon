package openclaw

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func parseNDJSONLine(line []byte) (agentbridge.RawEvent, bool) {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return agentbridge.RawEvent{}, false
	}

	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return agentbridge.RawEvent{}, false
	}

	event, _ := payload["event"].(string)
	if event == "" {
		return agentbridge.RawEvent{}, false
	}
	return rawPayloadEvent(
		agentbridge.RawSourceStdout,
		"ndjson:"+event,
		payload,
		trimmed,
	), true
}

func closeNDJSONStdout(buf []byte) []agentbridge.RawEvent {
	trimmed := bytes.TrimSpace(buf)
	if ev, ok := parseNDJSONLine(trimmed); ok {
		return []agentbridge.RawEvent{ev}
	}
	if len(trimmed) == 0 {
		return nil
	}
	return []agentbridge.RawEvent{
		rawBytesEvent(agentbridge.RawSourceClose, "malformed", trimmed),
	}
}
