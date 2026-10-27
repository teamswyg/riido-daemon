package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

type childExitRunning struct {
	stdout      chan []byte
	stderr      chan []byte
	exited      chan process.ExitStatus
	killStarted chan struct{}
	killFn      func(context.Context) error
}

func newChildExitRunning() *childExitRunning {
	return &childExitRunning{
		stdout: make(chan []byte), stderr: make(chan []byte),
		exited: make(chan process.ExitStatus, 1), killStarted: make(chan struct{}, 1),
	}
}

func (p *childExitRunning) Stdout() <-chan []byte             { return p.stdout }
func (p *childExitRunning) Stderr() <-chan []byte             { return p.stderr }
func (p *childExitRunning) Exited() <-chan process.ExitStatus { return p.exited }
func (p *childExitRunning) WriteStdin([]byte) error           { return nil }
func (p *childExitRunning) CloseStdin() error                 { return nil }
func (p *childExitRunning) Kill(ctx context.Context) error {
	select {
	case p.killStarted <- struct{}{}:
	default:
	}
	if p.killFn != nil {
		return p.killFn(ctx)
	}
	return nil
}

func newChildExitRunner(ctx context.Context, p *childExitRunning, timeout time.Duration) *sessionRunner {
	cfg := defaultSessionConfig(Config{
		TaskID: "assignment-1", RuntimeID: "runtime-1",
		Adapter:            &recordingAdapter{parser: &recordingParser{}},
		ProcessKillTimeout: timeout,
	})
	sess := newSessionHandle(cfg, p)
	r := newSessionRunner(ctx, cfg, sess, p)
	r.state.Terminal = true
	r.state.Result = agentbridge.Result{Status: agentbridge.ResultCompleted, Output: "provider result"}
	return r
}

func awaitChildExitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("child-exit fixture did not reach barrier")
	}
}

func emitChildExit(p *childExitRunning, code int) {
	close(p.stdout)
	close(p.stderr)
	p.exited <- process.ExitStatus{Code: code}
	close(p.exited)
}

type childExitStartFailure struct{}

func (childExitStartFailure) Start(context.Context, process.Command) (process.RunningProcess, error) {
	return nil, errors.New("fixture start failure")
}
