package main

import (
	"context"

	"github.com/teamswyg/riido-daemon/pkg/lifecycle"
)

func runDaemon(args []string) error {
	return runDaemonWithLifecycle(lifecycle.Background(), args)
}

// runDaemonWithContext lets tests cancel the foreground daemon by
// canceling ctx. It is a stdlib-compatibility wrapper around the daemon's
// named lifecycle context.
func runDaemonWithContext(ctx context.Context, args []string) error {
	return runDaemonWithLifecycle(lifecycle.FromContext(ctx), args)
}

func runDaemonWithLifecycle(ctx lifecycle.Context, args []string) error {
	if len(args) < 1 {
		printUsage()
		return daemonErrorf(ErrDaemonUsage, "run", "missing daemon subcommand")
	}
	if isHelpArg(args[0]) {
		printUsage()
		return nil
	}
	switch daemonCommand(args[0]) {
	case daemonCommandStart:
		return runDaemonStart(ctx, args[1:])
	case daemonCommandStatus:
		return runDaemonStatus(args[1:])
	case daemonCommandHealth:
		return runDaemonHealth(args[1:])
	case daemonCommandReady:
		return runDaemonReady(args[1:])
	case daemonCommandMetrics:
		return runDaemonMetrics(args[1:])
	case daemonCommandStop:
		return runDaemonStop(args[1:])
	case daemonCommandLogs:
		return runDaemonLogs(args[1:])
	default:
		printUsage()
		return daemonErrorf(ErrDaemonUsage, "run", "unknown daemon subcommand: %s", args[0])
	}
}
