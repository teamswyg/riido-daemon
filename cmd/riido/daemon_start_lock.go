package main

import (
	"context"
	"time"

	c9lock "github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func ensureDaemonStartLockAvailable(ctx lifecycle.Context, lockFile string) error {
	probeCtx, cancel := context.WithTimeout(ctx.Context(), 100*time.Millisecond)
	defer cancel()

	lock, err := c9lock.AcquireFile(probeCtx, lockFile)
	if err != nil {
		return daemonWrapf(ErrDaemonLock, "background.preflight-lock", err, "daemon already running or starting; singleton lock %s is held", lockFile)
	}
	if err := lock.Release(); err != nil {
		return daemonWrapf(ErrDaemonLock, "background.preflight-lock", err, "release daemon singleton probe lock %s", lockFile)
	}
	return nil
}
