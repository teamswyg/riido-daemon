package session

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

type recordingParser struct {
	stdoutChunks [][]byte
	stderrChunks [][]byte
	closed       bool
}

func (p *recordingParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.stdoutChunks = append(p.stdoutChunks, append([]byte(nil), chunk...))
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}

func (p *recordingParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	p.stderrChunks = append(p.stderrChunks, append([]byte(nil), chunk...))
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStderr, Type: "stderr-chunk", Bytes: chunk}}, nil
}

func (p *recordingParser) Close() ([]agentbridge.RawEvent, error) {
	p.closed = true
	return nil, nil
}
