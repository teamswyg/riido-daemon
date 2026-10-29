package session

import (
	"context"
	"time"

	"github.com/teamswyg/riido-daemon/internal/process"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func killProcess(ctx context.Context, proc process.RunningProcess, timeout time.Duration) error {
	killCtx, cancel := processKillContext(ctx, timeout)
	defer cancel()
	return proc.Kill(killCtx)
}

func processKillContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultProcessKillTimeout
	}
	level := lifecycle.FromContext(ctx).ShutdownLevel()
	killCtx, cancel := lifecycle.DetachedShutdown(lifecycle.NormalizeShutdownLevel(level), timeout)
	return killCtx.Context(), cancel
}
