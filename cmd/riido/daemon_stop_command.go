package main

import (
	"time"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func runDaemonStop(args []string) error {
	flags, err := parseDaemonStopFlags(args)
	if isCLIHelp(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if flags.socket == "" && flags.pidFile == "" {
		return daemonErrorf(ErrDaemonUsage, "stop.parse-flags", "daemon stop requires at least one of --socket or --pid-file")
	}
	return stopDaemonWithFlags(flags)
}

func daemonStopLevel(force bool) lifecycle.ShutdownLevel {
	if force {
		return lifecycle.ShutdownForced
	}
	return lifecycle.ShutdownGraceful
}

func stopDaemonWithFlags(flags daemonStopFlags) error {
	timeout := time.Duration(flags.timeoutSeconds) * time.Second
	level := daemonStopLevel(flags.force)
	if flags.socket != "" && tryShutdownViaSocket(flags.socket, timeout, level) {
		return nil
	}
	if flags.pidFile == "" {
		return daemonErrorf(ErrDaemonSocket, "stop.socket-fallback", "daemon stop: socket %s did not respond and --pid-file is not provided", flags.socket)
	}
	return stopViaPIDFile(flags.pidFile, timeout, level)
}
