package openclaw

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (p *parser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.fullStdoutBuf = append(p.fullStdoutBuf, chunk...)
	p.ndjsonLineBuf = append(p.ndjsonLineBuf, chunk...)

	var out []agentbridge.RawEvent
	for {
		line, ok := nextParserLine(&p.ndjsonLineBuf)
		if !ok {
			return out, nil
		}
		if ev, ok := parseNDJSONLine(line); ok {
			p.emittedNDJSON = true
			out = append(out, ev)
		}
	}
}
