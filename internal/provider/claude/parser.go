package claude

import (
	"bytes"
	"encoding/json"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

// MaxLineBytes is the largest stream-json line we will accept on either
// stdout or stderr. Claude's stream-json can carry very large tool results; the
// public provider-runtime SSOT keeps the adapter bounded at 10 MB.
const MaxLineBytes = 10 * 1024 * 1024

// stdoutStreamPrefixes / stderrStreamPrefixes are wrapper-script lead-ins
// that some launchers prepend to a line (e.g. when the daemon piped
// stderr through a script that echoed "stderr: ..."). The parser
// strips them so the translator sees the clean payload.
var stdoutStreamPrefixes = []string{"stdout: ", "STDOUT: "}
var stderrStreamPrefixes = []string{"stderr: ", "STDERR: "}

// parser is the Claude stream-json line scanner. State is owned by the
// single goroutine that calls FeedStdout/FeedStderr/Close. No mutex.
type parser struct {
	stdoutBuf []byte
	stderrBuf []byte
}

// NewParser returns an agentbridge.Parser for Claude's stream-json
// output. The returned parser is NOT safe for concurrent use — it is
// designed to be owned by a single SessionActor goroutine.
func NewParser() agentbridge.Parser {
	return &parser{}
}

func (p *parser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stdoutBuf, chunk, agentbridge.RawSourceStdout, stdoutStreamPrefixes, true)
}

func (p *parser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stderrBuf, chunk, agentbridge.RawSourceStderr, stderrStreamPrefixes, false)
}

func (p *parser) Close() ([]agentbridge.RawEvent, error) {
	var out []agentbridge.RawEvent
	if len(p.stdoutBuf) > 0 {
		ev, ok := p.parseLine(p.stdoutBuf, agentbridge.RawSourceClose, stdoutStreamPrefixes, true)
		p.stdoutBuf = nil
		if ok {
			out = append(out, ev)
		}
	}
	if len(p.stderrBuf) > 0 {
		ev, ok := p.parseLine(p.stderrBuf, agentbridge.RawSourceClose, stderrStreamPrefixes, false)
		p.stderrBuf = nil
		if ok {
			out = append(out, ev)
		}
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
		ev, ok := p.parseLine(line, source, prefixes, parseJSON)
		if ok {
			out = append(out, ev)
		}
	}
	if len(*buf) > MaxLineBytes {
		// Drop the over-long fragment to avoid unbounded growth while still
		// surfacing a malformed event so the translator sees something happened.
		bad := *buf
		*buf = nil
		out = append(out, agentbridge.RawEvent{
			Source: source,
			Type:   "malformed",
			Bytes:  append([]byte(nil), bad[:1024]...), // sample for diagnostics
		})
	}
	return out, nil
}

func (p *parser) parseLine(line []byte, source agentbridge.RawSource, prefixes []string, parseJSON bool) (agentbridge.RawEvent, bool) {
	// Trim CR (CRLF tolerance) and strip wrapper prefix.
	if n := len(line); n > 0 && line[n-1] == '\r' {
		line = line[:n-1]
	}
	for _, prefix := range prefixes {
		if bytes.HasPrefix(line, []byte(prefix)) {
			line = line[len(prefix):]
			break
		}
	}
	trimmed := bytes.TrimSpace(line)
	if len(trimmed) == 0 {
		return agentbridge.RawEvent{}, false
	}

	ev := agentbridge.RawEvent{
		Source: source,
		Bytes:  append([]byte(nil), trimmed...),
	}
	if !parseJSON {
		ev.Type = "stderr"
		return ev, true
	}

	// Stdout: try JSON first.
	if trimmed[0] == '{' || trimmed[0] == '[' {
		var payload map[string]any
		if err := json.Unmarshal(trimmed, &payload); err == nil {
			ev.Payload = payload
			if t, ok := payload["type"].(string); ok {
				ev.Type = t
			} else {
				ev.Type = "unknown"
			}
			return ev, true
		}
	}
	ev.Type = "malformed"
	return ev, true
}
