package supervisor

import "github.com/teamswyg/riido-daemon/internal/agentbridge"

type stubParser struct{}

func (p *stubParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	if string(chunk) == "event" {
		return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "event", Bytes: chunk}}, nil
	}
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, nil
}

func (p *stubParser) FeedStderr([]byte) ([]agentbridge.RawEvent, error) {
	return nil, nil
}

func (p *stubParser) Close() ([]agentbridge.RawEvent, error) {
	return nil, nil
}
