package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

func (p *parser) Close() ([]agentbridge.RawEvent, error) {
	var out []agentbridge.RawEvent
	out = p.appendCloseLine(out, &p.stdoutBuf, agentbridge.RawSourceClose, stdoutStreamPrefixes, true)
	out = p.appendCloseLine(out, &p.stderrBuf, agentbridge.RawSourceStderr, stderrStreamPrefixes, false)
	return out, nil
}

func (p *parser) appendCloseLine(
	out []agentbridge.RawEvent,
	buf *[]byte,
	source agentbridge.RawSource,
	prefixes []string,
	parseJSON bool,
) []agentbridge.RawEvent {
	if len(*buf) == 0 {
		return out
	}
	ev, ok := p.parseLine(*buf, source, prefixes, parseJSON)
	*buf = nil
	if !ok {
		return out
	}
	return append(out, ev)
}
