package codex

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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
