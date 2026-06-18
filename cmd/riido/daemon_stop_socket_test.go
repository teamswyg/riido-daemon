package main

import (
	"net"
	"testing"
	"time"
)

func TestDaemonStopViaSocket(t *testing.T) {
	paths := newDaemonBackgroundPaths(t)
	if err := run([]string{
		"daemon", "start",
		"--socket", paths.socket,
		"--pid-file", paths.pid,
		"--lock-file", paths.lock,
	}); err != nil {
		t.Fatalf("start: %v", err)
	}

	if err := run([]string{
		"daemon", "stop",
		"--socket", paths.socket,
		"--timeout-seconds", "5",
	}); err != nil {
		t.Fatalf("stop via socket: %v", err)
	}
	assertDaemonSocketClosed(t, paths.socket)
}

func assertDaemonSocketClosed(t *testing.T, socket string) {
	t.Helper()
	if _, err := net.DialTimeout("unix", socket, 200*time.Millisecond); err == nil {
		t.Fatal("daemon still responding after stop")
	}
}
