package claude

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

// parser is the Claude stream-json line scanner. State is owned by the
// single goroutine that calls FeedStdout/FeedStderr/Close. No mutex.
type parser struct {
	stdoutBuf []byte
	stderrBuf []byte
}

// NewParser returns an agentbridge.Parser for Claude's stream-json output.
// The returned parser is not safe for concurrent use.
func NewParser() agentbridge.Parser {
	return &parser{}
}

func (p *parser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stdoutBuf, chunk, agentbridge.RawSourceStdout, stdoutStreamPrefixes, true)
}

func (p *parser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	return p.feed(&p.stderrBuf, chunk, agentbridge.RawSourceStderr, stderrStreamPrefixes, false)
}
