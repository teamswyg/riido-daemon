package session

import (
	"context"

	"github.com/teamswyg/riido-daemon/internal/process"
)

func (r *sessionRunner) finishProcess() bool {
	ctx, cancel := processKillContext(r.ctx, r.cfg.ProcessKillTimeout)
	defer cancel()
	observed := make(chan bool, 1)
	// Drain concurrently: a child Wait can depend on both pipe readers
	// completing. Kill and observation share the same detached deadline.
	go func() {
		observed <- observeChildExit(ctx, r.stdoutCh, r.stderrCh, r.exitedCh, r.deferredExit != nil)
	}()
	_ = r.proc.Kill(ctx)
	return <-observed
}

func observeChildExit(ctx context.Context, stdout, stderr <-chan []byte, exited <-chan process.ExitStatus, observed bool) bool {
	if observed {
		exited = nil
	}
	for stdout != nil || stderr != nil || exited != nil {
		if ctx.Err() != nil {
			return observed
		}
		select {
		case <-ctx.Done():
			return observed
		case _, ok := <-stdout:
			if !ok {
				stdout = nil
			}
		case _, ok := <-stderr:
			if !ok {
				stderr = nil
			}
		case _, ok := <-exited:
			if ok {
				observed = true
			}
			exited = nil
		}
	}
	return observed
}
