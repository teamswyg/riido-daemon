package openclaw

import (
	"bytes"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func (p *parser) Close() ([]agentbridge.RawEvent, error) {
	var out []agentbridge.RawEvent
	out = append(out, p.closeStdout()...)
	out = append(out, p.closeStderr()...)
	return out, nil
}

func (p *parser) closeStdout() []agentbridge.RawEvent {
	trimmed := bytes.TrimSpace(p.fullStdoutBuf)
	p.fullStdoutBuf = nil
	if len(trimmed) == 0 {
		return nil
	}
	if p.emittedNDJSON {
		return closeNDJSONStdout(p.ndjsonLineBuf)
	}
	return closeFullResultStdout(trimmed)
}

func (p *parser) closeStderr() []agentbridge.RawEvent {
	rem := bytes.TrimSpace(p.stderrBuf)
	p.stderrBuf = nil
	if len(rem) == 0 {
		return nil
	}
	return []agentbridge.RawEvent{
		rawBytesEvent(agentbridge.RawSourceClose, "stderr", rem),
	}
}
