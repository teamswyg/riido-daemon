package cursor

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *parser) parseLine(line []byte, source agentbridge.RawSource, prefixes []string, parseJSON bool) (agentbridge.RawEvent, bool) {
	trimmed := trimLine(line, prefixes)
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
	return parseJSONLine(ev, trimmed)
}

func parseJSONLine(ev agentbridge.RawEvent, trimmed []byte) (agentbridge.RawEvent, bool) {
	var m map[string]any
	if err := json.Unmarshal(trimmed, &m); err != nil {
		ev.Type = "malformed"
		return ev, true
	}
	ev.Payload = m
	if t, ok := m["type"].(string); ok {
		ev.Type = t
	} else {
		ev.Type = "unknown"
	}
	return ev, true
}

func trimLine(line []byte, prefixes []string) []byte {
	if n := len(line); n > 0 && line[n-1] == '\r' {
		line = line[:n-1]
	}
	for _, pre := range prefixes {
		if bytes.HasPrefix(line, []byte(pre)) {
			line = line[len(pre):]
			break
		}
	}
	return bytes.TrimSpace(line)
}
