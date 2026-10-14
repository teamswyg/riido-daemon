package main

import "github.com/teamswyg/riido-daemon/pkg/lifecycle"

// runDaemonStartBackground forks the same binary in foreground mode and waits
// until the child accepts the daemon socket.
func runDaemonStartBackground(ctx lifecycle.Context, flags startFlags) error {
	if err := ensureDaemonStartLockAvailable(ctx, flags.lockFile); err != nil {
		return err
	}
	cmd, devNull, err := prepareDaemonBackgroundChild(flags)
	if err != nil {
		return err
	}
	defer func() { _ = devNull.Close() }()

	if err := cmd.Start(); err != nil {
		return daemonWrapf(ErrDaemonProcess, "background.spawn", err, "spawn daemon child")
	}
	if err := waitForDaemonChildReady(flags, cmd); err != nil {
		_ = cmd.Process.Kill()
		return err
	}
	return nil
}
