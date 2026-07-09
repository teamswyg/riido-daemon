package session

import (
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
)

func TestSessionStdoutParserErrorIsEmitted(t *testing.T) {
	parser := &edgeParser{stdoutErr: errors.New("bad stdout")}
	started := startEdgeSession(t, "task-stdout-parse-error", newEdgeAdapter(parser, nil))
	errCh := watchErrorText(started.session)
	started.running.EmitStdout([]byte("broken"))
	assertErrorContains(t, errCh, "bad stdout")
	started.session.Cancel(nil)
	_ = waitResult(t, started.session, time.Second)
}

func TestSessionStderrParserErrorDoesNotTerminate(t *testing.T) {
	called := make(chan struct{}, 1)
	parser := &edgeParser{stderrErr: errors.New("bad stderr"), stderrCalled: called}
	started := startEdgeSession(t, "task-stderr-parse-error", newEdgeAdapter(parser, nil))
	started.running.EmitStderr([]byte("noise"))
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("stderr parser was not called")
	}
	started.session.Cancel(errors.New("done"))
	res := waitResult(t, started.session, time.Second)
	if res.Status != agentbridge.ResultCancelled {
		t.Fatalf("status=%s", res.Status)
	}
}

func TestSessionTranslateErrorIsEmitted(t *testing.T) {
	parser := &edgeParser{}
	translate := func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) {
		return nil, nil, errors.New("translate failed")
	}
	started := startEdgeSession(t, "task-translate-error", newEdgeAdapter(parser, translate))
	errCh := watchErrorText(started.session)
	started.running.EmitStdout([]byte("raw"))
	assertErrorContains(t, errCh, "translate failed")
	started.session.Cancel(nil)
	_ = waitResult(t, started.session, time.Second)
}
