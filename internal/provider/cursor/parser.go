package cursor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

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
	out := p.closeBuffer(&p.stdoutBuf, agentbridge.RawSourceClose, stdoutPrefixes, true)
	out = append(out, p.closeBuffer(&p.stderrBuf, agentbridge.RawSourceClose, stderrPrefixes, false)...)
	return out, nil
}

func (p *parser) closeBuffer(buf *[]byte, source agentbridge.RawSource, prefixes []string, parseJSON bool) []agentbridge.RawEvent {
	if len(*buf) == 0 {
		return nil
	}
	defer func() { *buf = nil }()
	if ev, ok := p.parseLine(*buf, source, prefixes, parseJSON); ok {
		return []agentbridge.RawEvent{ev}
	}
	return nil
}
