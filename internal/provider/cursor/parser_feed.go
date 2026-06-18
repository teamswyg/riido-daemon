package cursor

import (
	"bytes"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *parser) feed(buf *[]byte, chunk []byte, source agentbridge.RawSource, prefixes []string, parseJSON bool) ([]agentbridge.RawEvent, error) {
	*buf = append(*buf, chunk...)
	var out []agentbridge.RawEvent
	for {
		idx := bytes.IndexByte(*buf, '\n')
		if idx < 0 {
			break
		}
		line := (*buf)[:idx]
		*buf = (*buf)[idx+1:]
		if ev, ok := p.parseLine(line, source, prefixes, parseJSON); ok {
			out = append(out, ev)
		}
	}
	return append(out, oversizedLineEvent(buf, source)...), nil
}

func oversizedLineEvent(buf *[]byte, source agentbridge.RawSource) []agentbridge.RawEvent {
	if len(*buf) <= MaxLineBytes {
		return nil
	}
	bad := *buf
	*buf = nil
	return []agentbridge.RawEvent{{Source: source, Type: "malformed", Bytes: append([]byte(nil), bad[:1024]...)}}
}
