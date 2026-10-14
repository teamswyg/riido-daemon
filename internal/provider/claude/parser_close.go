package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (p *parser) Close() ([]agentbridge.RawEvent, error) {
	var out []agentbridge.RawEvent
	out = p.appendCloseLine(out, &p.stdoutBuf, stdoutStreamPrefixes, true)
	out = p.appendCloseLine(out, &p.stderrBuf, stderrStreamPrefixes, false)
	return out, nil
}

func (p *parser) appendCloseLine(
	out []agentbridge.RawEvent,
	buf *[]byte,
	prefixes []string,
	parseJSON bool,
) []agentbridge.RawEvent {
	if len(*buf) == 0 {
		return out
	}
	ev, ok := p.parseLine(*buf, agentbridge.RawSourceClose, prefixes, parseJSON)
	*buf = nil
	if !ok {
		return out
	}
	return append(out, ev)
}
