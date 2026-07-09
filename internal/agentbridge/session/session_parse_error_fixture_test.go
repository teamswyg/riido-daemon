package session

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

type edgeAdapter struct {
	parser    agentbridge.Parser
	translate func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error)
}

func newEdgeAdapter(
	parser agentbridge.Parser,
	translate func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error),
) *edgeAdapter {
	return &edgeAdapter{parser: parser, translate: translate}
}

func (a *edgeAdapter) Name() string { return "edge" }

func (a *edgeAdapter) Detect(context.Context, agentbridge.DetectEnv) (agentbridge.DetectResult, error) {
	return agentbridge.DetectResult{Available: true}, nil
}

func (a *edgeAdapter) BuildStart(agentbridge.StartRequest) (agentbridge.StartCommand, error) {
	return agentbridge.StartCommand{}, nil
}

func (a *edgeAdapter) NewParser() agentbridge.Parser { return a.parser }

func (a *edgeAdapter) Translate(raw agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
	if a.translate != nil {
		return a.translate(raw)
	}
	return nil, nil, nil
}

func (a *edgeAdapter) BlockedArgs() []string { return nil }

func assertErrorContains(t *testing.T, errCh <-chan string, want string) {
	t.Helper()
	select {
	case got := <-errCh:
		if !strings.Contains(got, want) {
			t.Fatalf("error=%q", got)
		}
	case <-time.After(time.Second):
		t.Fatalf("missing error containing %q", want)
	}
}

type edgeParser struct {
	stdoutErr    error
	stderrErr    error
	stderrCalled chan<- struct{}
}

func (p *edgeParser) FeedStdout(chunk []byte) ([]agentbridge.RawEvent, error) {
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStdout, Type: "chunk", Bytes: chunk}}, p.stdoutErr
}

func (p *edgeParser) FeedStderr(chunk []byte) ([]agentbridge.RawEvent, error) {
	if p.stderrCalled != nil {
		p.stderrCalled <- struct{}{}
	}
	return []agentbridge.RawEvent{{Source: agentbridge.RawSourceStderr, Type: "chunk", Bytes: chunk}}, p.stderrErr
}

func (p *edgeParser) Close() ([]agentbridge.RawEvent, error) { return nil, nil }
