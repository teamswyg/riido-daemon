package codex

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *parser) parseLine(
	line []byte,
	source agentbridge.RawSource,
	parseJSON bool,
) (agentbridge.RawEvent, bool) {
	if n := len(line); n > 0 && line[n-1] == '\r' {
		line = line[:n-1]
	}
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return agentbridge.RawEvent{}, false
	}
	ev := agentbridge.RawEvent{Source: source, Bytes: append([]byte(nil), trimmed...)}
	if !parseJSON {
		ev.Type = "stderr"
		return ev, true
	}
	if trimmed[0] != '{' {
		ev.Type = "malformed"
		return ev, true
	}
	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		ev.Type = "malformed"
		return ev, true
	}
	ev.Payload = payload
	ev.Type = classifyJSONRPC(payload)
	return ev, true
}
