package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/process"
)

type blockingKillRunning struct {
	stdout  chan []byte
	stderr  chan []byte
	exited  chan process.ExitStatus
	kill    chan struct{}
	unblock chan struct{}
}

func newBlockingKillRunning() *blockingKillRunning {
	return &blockingKillRunning{
		stdout:  make(chan []byte),
		stderr:  make(chan []byte),
		exited:  make(chan process.ExitStatus),
		kill:    make(chan struct{}, 1),
		unblock: make(chan struct{}),
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
	select {
	case <-r.unblock:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *blockingKillRunning) KillRecv() <-chan struct{} {
	return r.kill
}
