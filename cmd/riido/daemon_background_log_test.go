package main

import (
	"bytes"
	"os"
	"testing"
	"time"
)

func TestDaemonBackgroundLogFileReceivesChildOutput(t *testing.T) {
	paths := newDaemonBackgroundPaths(t)
	if err := run(paths.startArgs()); err != nil {
		t.Fatalf("start: %v", err)
	}
	cleanupBackgroundDaemon(t, paths)

	_, _ = runCapturingStdout(t, func() error {
		return run([]string{"daemon", "status", "--socket", paths.socket})
	})
	assertBackgroundLogContainsDaemon(t, paths.log)
}

func assertBackgroundLogContainsDaemon(t *testing.T, logPath string) {
	t.Helper()
	end := time.Now().Add(2 * time.Second)
	for time.Now().Before(end) {
		body, err := os.ReadFile(logPath)
		if err == nil && bytes.Contains(body, []byte("daemon")) {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	body, _ := os.ReadFile(logPath)
	t.Fatalf("log file %s did not receive daemon output. content:\n%s", logPath, body)
}
