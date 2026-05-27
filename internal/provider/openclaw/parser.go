package openclaw

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// parser handles OpenClaw's two output modes:
//
//  1. Full JSON result on stdout: one object emitted once at the end of the
//     run, either compact or pretty-printed. The parser buffers stdout until
//     Close, then tries to decode the whole buffer as a single JSON object →
//     RawEvent of Type "full_result".
//
//  2. NDJSON streaming: line-delimited JSON events. When we see a
//     line-terminated JSON object during Feed, we emit a RawEvent of
//     Type "ndjson:<event>" eagerly.
//
// The mode is detected per-line: any line that parses as JSON during
// FeedStdout is treated as NDJSON. If no lines are emitted during the
// run (i.e. all output came in one chunk with no trailing newline) the
// buffer is decoded once on Close.
//
// This matches the “안정 호환은 full JSON 우선, 실패 시 NDJSON fallback”
// pattern from spec §3.3.
type parser struct {
	fullStdoutBuf []byte
	ndjsonLineBuf []byte
	stderrBuf     []byte
	emittedNDJSON bool
}

func NewParser() agentbridge.Parser { return &parser{} }

func (p *parser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.fullStdoutBuf = append(p.fullStdoutBuf, chunk...)
	p.ndjsonLineBuf = append(p.ndjsonLineBuf, chunk...)
	var out []agentbridge.RawEvent
	for {
		idx := bytes.IndexByte(p.ndjsonLineBuf, '\n')
		if idx < 0 {
			break
		}
		line := p.ndjsonLineBuf[:idx]
		p.ndjsonLineBuf = p.ndjsonLineBuf[idx+1:]
		if ev, ok := parseNDJSONLine(line); ok {
			p.emittedNDJSON = true
			out = append(out, ev)
		}
	}
	return out, nil
}

func (p *parser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.stderrBuf = append(p.stderrBuf, chunk...)
	var out []agentbridge.RawEvent
	for {
		idx := bytes.IndexByte(p.stderrBuf, '\n')
		if idx < 0 {
			break
		}
		line := p.stderrBuf[:idx]
		p.stderrBuf = p.stderrBuf[idx+1:]
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			continue
		}
		out = append(out, agentbridge.RawEvent{Source: agentbridge.RawSourceStderr, Type: "stderr", Bytes: append([]byte(nil), trimmed...)})
	}
	return out, nil
}

func (p *parser) Close() ([]agentbridge.RawEvent, error) {
	var out []agentbridge.RawEvent
	trimmed := bytes.TrimSpace(p.fullStdoutBuf)
	p.fullStdoutBuf = nil
	if len(trimmed) > 0 {
		if p.emittedNDJSON {
			// NDJSON mode: trailing fragment without newline. Try a
			// single ndjson decode; if it fails, surface as malformed.
			if ev, ok := parseNDJSONLine(bytes.TrimSpace(p.ndjsonLineBuf)); ok {
				out = append(out, ev)
			} else if len(bytes.TrimSpace(p.ndjsonLineBuf)) > 0 {
				out = append(out, agentbridge.RawEvent{Source: agentbridge.RawSourceClose, Type: "malformed", Bytes: append([]byte(nil), bytes.TrimSpace(p.ndjsonLineBuf)...)})
			}
		} else {
			// Full-result mode: decode the whole buffer.
			var m map[string]any
			if err := json.Unmarshal(trimmed, &m); err != nil {
				out = append(out, agentbridge.RawEvent{Source: agentbridge.RawSourceClose, Type: "malformed", Bytes: append([]byte(nil), trimmed...)})
			} else {
				out = append(out, agentbridge.RawEvent{Source: agentbridge.RawSourceClose, Type: "full_result", Payload: m, Bytes: append([]byte(nil), trimmed...)})
			}
		}
	}
	if rem := bytes.TrimSpace(p.stderrBuf); len(rem) > 0 {
		out = append(out, agentbridge.RawEvent{Source: agentbridge.RawSourceClose, Type: "stderr", Bytes: append([]byte(nil), rem...)})
		p.stderrBuf = nil
	}
	return out, nil
}

func parseNDJSONLine(line []byte) (agentbridge.RawEvent, bool) {
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return agentbridge.RawEvent{}, false
	}
	var m map[string]any
	if err := json.Unmarshal(trimmed, &m); err != nil {
		return agentbridge.RawEvent{}, false
	}
	event, _ := m["event"].(string)
	if event == "" {
		return agentbridge.RawEvent{}, false
	}
	return agentbridge.RawEvent{
		Source:  agentbridge.RawSourceStdout,
		Type:    "ndjson:" + event,
		Payload: m,
		Bytes:   append([]byte(nil), trimmed...),
	}, true
}
