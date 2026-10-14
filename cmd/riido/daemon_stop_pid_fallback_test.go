package main

import (
	"path/filepath"
	"testing"
)

func TestDaemonStopSocketFallsBackToPID(t *testing.T) {
	paths := newDaemonBackgroundPaths(t)
	if err := run([]string{
		"daemon", "start",
		"--socket", paths.socket,
		"--pid-file", paths.pid,
		"--lock-file", paths.lock,
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	cleanupBackgroundDaemon(t, paths)

	wrongSocket := filepath.Join(t.TempDir(), "wrong.sock")
	if err := run([]string{
		"daemon", "stop",
		"--socket", wrongSocket,
		"--pid-file", paths.pid,
		"--timeout-seconds", "3",
	}); err != nil {
		t.Fatalf("stop with wrong socket but right pid-file: %v", err)
	}
	assertDaemonSocketClosed(t, paths.socket)
}
