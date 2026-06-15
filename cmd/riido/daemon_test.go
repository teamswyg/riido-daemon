package main

import (
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// daemonSocketPath returns a short Unix socket path. macOS SUN_PATH is
// capped at 104 bytes, so we cannot use t.TempDir() (which on macOS
// produces ~100+ char paths). Instead we allocate a short
// /tmp/riido-<random>.sock and remove it via t.Cleanup.
func daemonSocketPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("/tmp", "riido-*.sock")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	path := f.Name()
	_ = f.Close()
	_ = os.Remove(path) // listen will recreate
	t.Cleanup(func() { _ = os.Remove(path) })
	return path
}

func daemonLockPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "agentd.lock")
}

// dialDaemon waits up to deadline for the daemon's local socket to accept.
func dialDaemon(t *testing.T, sock string, deadline time.Duration) {
	t.Helper()
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		conn, err := net.Dial("unix", sock)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("daemon socket %s never became reachable", sock)
}

// runCapturingStdout redirects os.Stdout for the duration of fn and
// returns what was written. fn's error is propagated.
func runCapturingStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	cmdErr := fn()
	_ = w.Close()
	<-done
	os.Stdout = orig
	return buf.String(), cmdErr
}
