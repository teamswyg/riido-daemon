package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (p *parser) parseLine(
	line []byte,
	source agentbridge.RawSource,
	prefixes []string,
	parseJSON bool,
) (agentbridge.RawEvent, bool) {
	trimmed := normalizeStreamLine(line, prefixes)
	if len(trimmed) == 0 {
		return agentbridge.RawEvent{}, false
	}

	ev := agentbridge.RawEvent{
		Source: source,
		Bytes:  append([]byte(nil), trimmed...),
	}
	if !parseJSON {
		ev.Type = "stderr"
		return ev, true
	}
	return decodeJSONRawEvent(ev, trimmed), true
}
