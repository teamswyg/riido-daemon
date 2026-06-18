package processexec

import (
	"context"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func (r *execRunning) Kill(ctx context.Context) error {
	lctx := lifecycle.StopContext(ctx)
	if lctx.ShutdownLevel().IsForced() {
		r.forceTerminateProcessGroup()
		r.cancel()
		return nil
	}
	r.gracefulTerminateProcessGroup()
	select {
	case <-r.done:
	case <-lctx.Done():
		r.forceTerminateProcessGroup()
	}
	r.cancel()
	return nil
}

func (r *execRunning) killOnContext(ctx context.Context) {
	select {
	case <-ctx.Done():
		_ = r.Kill(ctx)
	case <-r.done:
	}
}

func (r *execRunning) gracefulTerminateProcessGroup() {
	r.termOnce.Do(func() {
		gracefulTerminateCommand(r.cmd)
	})
}

func (r *execRunning) forceTerminateProcessGroup() {
	r.forceOnce.Do(func() {
		forceTerminateCommand(r.cmd)
	})
}
