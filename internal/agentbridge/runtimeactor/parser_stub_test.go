package runtimeactor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

type stubParser struct{}

func (p *stubParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{{
		Source: agentbridge.RawSourceStdout,
		Type:   "chunk",
		Bytes:  chunk,
	}}, nil
}

func (p *stubParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	return nil, nil
}

func (p *stubParser) Close() ([]agentbridge.RawEvent, error) { return nil, nil }
