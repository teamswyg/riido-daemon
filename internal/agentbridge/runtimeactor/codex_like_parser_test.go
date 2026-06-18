package runtimeactor

import (
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// lineJSONParser turns each newline-terminated JSON line into a RawEvent.
type lineJSONParser struct{ buf []byte }

func (p *lineJSONParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.buf = append(p.buf, chunk...)
	var out []agentbridge.RawEvent
	for {
		line, ok := p.nextLine()
		if !ok {
			return out, nil
		}
		if raw, ok := rawJSONLine(line); ok {
			out = append(out, raw)
		}
	}
}

func (p *lineJSONParser) FeedStderr(_ []byte) ([]agentbridge.RawEvent, error) { return nil, nil }
func (p *lineJSONParser) Close() ([]agentbridge.RawEvent, error)              { return nil, nil }

func (p *lineJSONParser) nextLine() ([]byte, bool) {
	for i, b := range p.buf {
		if b == '\n' {
			line := p.buf[:i]
			p.buf = p.buf[i+1:]
			return line, true
		}
	}
	return nil, false
}

func rawJSONLine(line []byte) (agentbridge.RawEvent, bool) {
	if len(line) == 0 {
		return agentbridge.RawEvent{}, false
	}
	var payload map[string]any
	if err := json.Unmarshal(line, &payload); err != nil {
		return agentbridge.RawEvent{}, false
	}
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    jsonRPCFrameType(payload),
		Payload: payload,
		Bytes:   line,
	}, true
}
