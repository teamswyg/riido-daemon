package cursor

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

const MaxLineBytes = 10 * 1024 * 1024

// Cursor wrappers sometimes prefix output lines with "stdout:" / "stderr:"
// from upstream scripts. We strip those before parsing JSON.
var (
	stdoutPrefixes = []string{"stdout: ", "STDOUT: "}
	stderrPrefixes = []string{"stderr: ", "STDERR: "}
)

type parser struct {
	stdoutBuf []byte
	stderrBuf []byte
}

func NewParser() agentbridge.Parser { return &parser{} }

func (p *parser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stdoutBuf, chunk, agentbridge.RawSourceStdout, stdoutPrefixes, true)
}

func (p *parser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stderrBuf, chunk, agentbridge.RawSourceStderr, stderrPrefixes, false)
}

func (p *parser) Close() ([]agentbridge.RawEvent, error) {
	var out []agentbridge.RawEvent
	if len(p.stdoutBuf) > 0 {
		if ev, ok := p.parseLine(p.stdoutBuf, agentbridge.RawSourceClose, stdoutPrefixes, true); ok {
			out = append(out, ev)
		}
		p.stdoutBuf = nil
	}
	if len(p.stderrBuf) > 0 {
		if ev, ok := p.parseLine(p.stderrBuf, agentbridge.RawSourceClose, stderrPrefixes, false); ok {
			out = append(out, ev)
		}
		p.stderrBuf = nil
	}
	return out, nil
}

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
	if len(*buf) > MaxLineBytes {
		bad := *buf
		*buf = nil
		out = append(out, agentbridge.RawEvent{Source: source, Type: "malformed", Bytes: append([]byte(nil), bad[:1024]...)})
	}
	return out, nil
}

func (p *parser) parseLine(line []byte, source agentbridge.RawSource, prefixes []string, parseJSON bool) (agentbridge.RawEvent, bool) {
	if n := len(line); n > 0 && line[n-1] == '\r' {
		line = line[:n-1]
	}
	for _, pre := range prefixes {
		if bytes.HasPrefix(line, []byte(pre)) {
			line = line[len(pre):]
			break
		}
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
	var m map[string]any
	if err := json.Unmarshal(trimmed, &m); err != nil {
		ev.Type = "malformed"
		return ev, true
	}
	ev.Payload = m
	if t, ok := m["type"].(string); ok {
		ev.Type = t
	} else {
		ev.Type = "unknown"
	}
	return ev, true
}
