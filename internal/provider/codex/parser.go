package codex

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// MaxLineBytes is the largest JSON-RPC frame we accept. Codex frames
// can carry large patch contents on file-change requests; 10 MB matches
// the per-line bound we use for stream-json elsewhere.
const MaxLineBytes = 10 * 1024 * 1024

// parser is the Codex JSON-RPC line scanner. Codex's app-server
// transport is LSP-style line-delimited JSON-RPC over stdio (one frame
// per line). Owner: a single SessionActor goroutine.
type parser struct {
	stdoutBuf []byte
	stderrBuf []byte
}

func NewParser() agentbridge.Parser { return &parser{} }

func (p *parser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stdoutBuf, chunk, agentbridge.RawSourceStdout, true)
}

func (p *parser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stderrBuf, chunk, agentbridge.RawSourceStderr, false)
}

func (p *parser) Close() ([]agentbridge.RawEvent, error) {
	var out []agentbridge.RawEvent
	if len(p.stdoutBuf) > 0 {
		if ev, ok := p.parseLine(p.stdoutBuf, agentbridge.RawSourceClose, true); ok {
			out = append(out, ev)
		}
		p.stdoutBuf = nil
	}
	if len(p.stderrBuf) > 0 {
		if ev, ok := p.parseLine(p.stderrBuf, agentbridge.RawSourceClose, false); ok {
			out = append(out, ev)
		}
		p.stderrBuf = nil
	}
	return out, nil
}

func (p *parser) feed(buf *[]byte, chunk []byte, source agentbridge.RawSource, parseJSON bool) ([]agentbridge.RawEvent, error) {
	*buf = append(*buf, chunk...)
	var out []agentbridge.RawEvent
	for {
		idx := bytes.IndexByte(*buf, '\n')
		if idx < 0 {
			break
		}
		line := (*buf)[:idx]
		*buf = (*buf)[idx+1:]
		if ev, ok := p.parseLine(line, source, parseJSON); ok {
			out = append(out, ev)
		}
	}
	if len(*buf) > MaxLineBytes {
		bad := *buf
		*buf = nil
		out = append(out, agentbridge.RawEvent{Source: source, Type: "malformed", Bytes: append([]byte(nil), bad[:1024]...)})
	}
	return out, nil
}

func (p *parser) parseLine(line []byte, source agentbridge.RawSource, parseJSON bool) (agentbridge.RawEvent, bool) {
	if n := len(line); n > 0 && line[n-1] == '\r' {
		line = line[:n-1]
	}
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return agentbridge.RawEvent{}, false
	}
	ev := agentbridge.RawEvent{Source: source, Bytes: append([]byte(nil), trimmed...)}
	if !parseJSON {
		ev.Type = "stderr"
		return ev, true
	}
	if trimmed[0] != '{' {
		ev.Type = "malformed"
		return ev, true
	}
	var payload map[string]any
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		ev.Type = "malformed"
		return ev, true
	}
	ev.Payload = payload
	ev.Type = classifyJSONRPC(payload)
	return ev, true
}

// classifyJSONRPC tags a JSON-RPC 2.0 frame:
//   - "notification:<method>" — method present, no id
//   - "server_request:<method>" — method present, id present
//   - "response" — result present, no method
//   - "error" — error present, no method
//   - "unknown" — none of the above
func classifyJSONRPC(payload map[string]any) string {
	method, hasMethod := payload["method"].(string)
	_, hasID := payload["id"]
	if hasMethod {
		if hasID {
			return "server_request:" + method
		}
		return "notification:" + method
	}
	if _, hasResult := payload["result"]; hasResult {
		return "response"
	}
	if _, hasError := payload["error"]; hasError {
		return "error"
	}
	return "unknown"
}
