package main

import (
	"path/filepath"
	"testing"
)

type daemonBackgroundPaths struct {
	socket string
	lock   string
	pid    string
	log    string
}

func newDaemonBackgroundPaths(t *testing.T) daemonBackgroundPaths {
	t.Helper()
	tempDir := t.TempDir()
	return daemonBackgroundPaths{
		socket: daemonSocketPath(t),
		lock:   daemonLockPath(t),
		pid:    filepath.Join(tempDir, "agentd.pid"),
		log:    filepath.Join(tempDir, "agentd.log"),
	}
}

func (p daemonBackgroundPaths) startArgs() []string {
	return []string{
		"daemon", "start",
		"--socket", p.socket,
		"--pid-file", p.pid,
		"--log-file", p.log,
		"--lock-file", p.lock,
	}
}

func (p daemonBackgroundPaths) stopArgs() []string {
	return []string{
		"daemon", "stop",
		"--socket", p.socket,
		"--pid-file", p.pid,
		"--timeout-seconds", "3",
	}
}

func cleanupBackgroundDaemon(t *testing.T, paths daemonBackgroundPaths) {
	t.Helper()
	t.Cleanup(func() { _ = run(paths.stopArgs()) })
}
