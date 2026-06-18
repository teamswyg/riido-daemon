package codex

import (
	"bytes"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *parser) feed(
	buf *[]byte,
	chunk []byte,
	source agentbridge.RawSource,
	parseJSON bool,
) ([]agentbridge.RawEvent, error) {
	*buf = append(*buf, chunk...)
	out := p.drainCompleteLines(buf, source, parseJSON)
	if len(*buf) > MaxLineBytes {
		bad := *buf
		*buf = nil
		out = append(out, agentbridge.RawEvent{
			Source: source,
			Type:   "malformed",
			Bytes:  append([]byte(nil), bad[:1024]...),
		})
	}
	return out, nil
}

func (p *parser) drainCompleteLines(
	buf *[]byte,
	source agentbridge.RawSource,
	parseJSON bool,
) []agentbridge.RawEvent {
	var out []agentbridge.RawEvent
	for {
		idx := bytes.IndexByte(*buf, '\n')
		if idx < 0 {
			return out
		}
		line := (*buf)[:idx]
		*buf = (*buf)[idx+1:]
		if ev, ok := p.parseLine(line, source, parseJSON); ok {
			out = append(out, ev)
		}
	}
}
