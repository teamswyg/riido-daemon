package runtimeactor

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/process"
)

type blockingKillProcess struct {
	running *blockingKillRunning
}

func newBlockingKillProcess() *blockingKillProcess {
	return &blockingKillProcess{running: newBlockingKillRunning()}
}

func (p *blockingKillProcess) Start(_ context.Context, _ process.Command) (process.RunningProcess, error) {
	return p.running, nil
}

func (p *blockingKillProcess) unblock() {
	close(p.running.unblock)
}
