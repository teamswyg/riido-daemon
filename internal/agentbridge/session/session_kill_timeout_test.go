package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/process"
)

func TestSessionTimeoutResultDoesNotWaitForeverForBlockingKill(t *testing.T) {
	proc := &blockingKillProcess{running: newBlockingKillRunning()}
	adapter := &recordingAdapter{
		name:        "fake",
		parser:      &recordingParser{},
		translateFn: func(agentbridge.RawEvent) ([]agentbridge.Event, []agentbridge.Command, error) { return nil, nil, nil },
	}
	sess, err := Start(context.Background(), Config{
		TaskID:             "task-blocking-kill-timeout",
		RuntimeID:          "rt-1",
		Adapter:            adapter,
		Process:            proc,
		Spawn:              process.Command{Executable: "fake"},
		HardTimeout:        10 * time.Millisecond,
		ProcessKillTimeout: 10 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	res := waitResult(t, sess, time.Second)
	if res.Status != agentbridge.ResultTimeout {
		t.Fatalf("result: %+v", res)
	}
	select {
	case <-proc.running.KillRecv():
	case <-time.After(time.Second):
		t.Fatal("session did not request provider kill")
	}
}

type blockingKillProcess struct {
	running *blockingKillRunning
}

func (p *blockingKillProcess) Start(context.Context, process.Command) (process.RunningProcess, error) {
	return p.running, nil
}

type blockingKillRunning struct {
	stdout chan []byte
	stderr chan []byte
	exited chan process.ExitStatus
	kill   chan struct{}
}

func newBlockingKillRunning() *blockingKillRunning {
	return &blockingKillRunning{
		stdout: make(chan []byte),
		stderr: make(chan []byte),
		exited: make(chan process.ExitStatus),
		kill:   make(chan struct{}, 2),
	}
}

func (r *blockingKillRunning) Stdout() <-chan []byte {
	return r.stdout
}

func (r *blockingKillRunning) Stderr() <-chan []byte {
	return r.stderr
}

func (r *blockingKillRunning) Exited() <-chan process.ExitStatus {
	return r.exited
}

func (r *blockingKillRunning) WriteStdin([]byte) error {
	return nil
}

func (r *blockingKillRunning) CloseStdin() error {
	return nil
}

func (r *blockingKillRunning) Kill(ctx context.Context) error {
	select {
	case r.kill <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return errors.New("blocking kill released by context")
}

func (r *blockingKillRunning) KillRecv() <-chan struct{} {
	return r.kill
}
