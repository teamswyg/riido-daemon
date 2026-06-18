package claude

import (
	"bytes"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *parser) feed(
	buf *[]byte,
	chunk []byte,
	source agentbridge.RawSource,
	prefixes []string,
	parseJSON bool,
) ([]agentbridge.RawEvent, error) {
	*buf = append(*buf, chunk...)
	var out []agentbridge.RawEvent
	for {
		idx := bytes.IndexByte(*buf, '\n')
		if idx < 0 {
			break
		}
		line := (*buf)[:idx]
		*buf = (*buf)[idx+1:]
		out = p.appendParsedLine(out, line, source, prefixes, parseJSON)
	}
	return appendOversizedFragment(out, buf, source), nil
}

func (p *parser) appendParsedLine(
	out []agentbridge.RawEvent,
	line []byte,
	source agentbridge.RawSource,
	prefixes []string,
	parseJSON bool,
) []agentbridge.RawEvent {
	ev, ok := p.parseLine(line, source, prefixes, parseJSON)
	if !ok {
		return out
	}
	return append(out, ev)
}

func appendOversizedFragment(
	out []agentbridge.RawEvent,
	buf *[]byte,
	source agentbridge.RawSource,
) []agentbridge.RawEvent {
	if len(*buf) <= MaxLineBytes {
		return out
	}
	bad := *buf
	*buf = nil
	return append(out, agentbridge.RawEvent{
		Source: source,
		Type:   "malformed",
		Bytes:  append([]byte(nil), bad[:1024]...),
	})
}
