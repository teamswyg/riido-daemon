package openclaw

import (
	"bytes"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *parser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.stderrBuf = append(p.stderrBuf, chunk...)

	var out []agentbridge.RawEvent
	for {
		line, ok := nextParserLine(&p.stderrBuf)
		if !ok {
			return out, nil
		}
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		out = append(out, rawBytesEvent(agentbridge.RawSourceStderr, "stderr", trimmed))
	}
}
