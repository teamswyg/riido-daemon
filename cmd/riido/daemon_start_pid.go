package main

import (
	"os"
	"strconv"
	"strings"
)

func ensureDaemonSocketOwnedByChild(flags startFlags, childPID int) error {
	if flags.pidFile == "" {
		return nil
	}
	raw, err := os.ReadFile(flags.pidFile)
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "background.verify-child-pid", err, "read daemon pid file")
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return daemonWrapf(ErrDaemonIO, "background.verify-child-pid", err, "parse daemon pid file")
	}
	if pid != childPID {
		return daemonErrorf(
			ErrDaemonLock,
			"background.verify-child-pid",
			"daemon socket %s is already served by pid %d; spawned child pid %d",
			flags.socket,
			pid,
			childPID,
		)
	}
	return nil
}
