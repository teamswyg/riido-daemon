package main

import (
	"net"
	"os/exec"
	"time"
)

func waitForDaemonChildReady(flags startFlags, cmd *exec.Cmd) error {
	exitCh := make(chan error, 1)
	go func() { exitCh <- cmd.Wait() }()

	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(50 * time.Millisecond)
	defer poll.Stop()

	for {
		select {
		case err := <-exitCh:
			return daemonWrapf(ErrDaemonProcess, "background.wait-ready", err, "daemon child exited before socket was ready")
		case <-deadline.C:
			_ = cmd.Process.Kill()
			return daemonErrorf(ErrDaemonSocket, "background.wait-ready", "daemon socket %s did not become ready within 15s", flags.socket)
		case <-poll.C:
			if !daemonSocketReady(flags.socket) {
				continue
			}
			if err := ensureDaemonSocketOwnedByChild(flags, cmd.Process.Pid); err != nil {
				_ = cmd.Process.Kill()
				return err
			}
			return nil
		}
	}
}

func daemonSocketReady(socket string) bool {
	conn, err := net.DialTimeout("unix", socket, 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
