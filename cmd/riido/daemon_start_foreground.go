package main

import (
	"context"
	"os"
	"time"

	c9lock "github.com/teamswyg/riido-daemon/internal/lock"
	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

const daemonForegroundLockAcquireTimeout = 100 * time.Millisecond

// runDaemonStartForeground is the in-process daemon path used by managed
// launchers and by the background wrapper after it re-invokes this binary.
func runDaemonStartForeground(ctx lifecycle.Context, flags startFlags) error {
	settings, err := loadDaemonSettings()
	if err != nil {
		return err
	}
	lockCtx, cancelLock := context.WithTimeout(ctx.Context(), daemonForegroundLockAcquireTimeout)
	defer cancelLock()
	lock, err := c9lock.AcquireFile(lockCtx, flags.lockFile)
	if err != nil {
		return daemonWrapf(ErrDaemonLock, "start.acquire-lock", err, "acquire daemon singleton lock %s", flags.lockFile)
	}
	defer releaseDaemonStartLock(lock)

	logSink, closeLog, err := openLogSink(flags.logFile)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "start.open-log", err, "open log sink")
	}
	defer closeLog()

	if flags.pidFile != "" {
		if err := writeDaemonPIDFiles(flags.pidFile, flags.socket); err != nil {
			return daemonWrapf(ErrDaemonIO, "start.write-pid", err, "write pid file")
		}
		defer removeDaemonPIDFiles(flags.pidFile)
	}

	logSink.Printf("daemon starting id=%s profile=%s socket=%s pid=%d", settings.DaemonID, settings.Profile, flags.socket, os.Getpid())
	return serveAgentDaemon(ctx, flags, settings, logSink)
}

func releaseDaemonStartLock(lock interface{ Release() error }) {
	if releaseErr := lock.Release(); releaseErr != nil {
		_, _ = os.Stderr.WriteString("riido daemon: release lock: " + releaseErr.Error() + "\n")
	}
}
