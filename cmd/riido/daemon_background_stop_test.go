package main

import (
	"context"
	"net"
	"path/filepath"
	"testing"
	"time"
)

// TestDaemonStopViaSocket: stop with --socket only (no --pid-file)
// must shut the daemon down via socket-shutdown request.
func TestDaemonStopViaSocket(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	pidPath := filepath.Join(t.TempDir(), "agentd.pid")

	if err := run([]string{
		"daemon", "start",
		"--socket", sock,
		"--pid-file", pidPath,
		"--lock-file", lockPath,
	}); err != nil {
		t.Fatalf("start: %v", err)
	}

	if err := run([]string{
		"daemon", "stop",
		"--socket", sock,
		"--timeout-seconds", "5",
	}); err != nil {
		t.Fatalf("stop via socket: %v", err)
	}

	// Socket must no longer accept connections.
	if _, err := net.DialTimeout("unix", sock, 200*time.Millisecond); err == nil {
		t.Fatal("daemon still responding after socket-shutdown stop")
	}
}

// TestDaemonStopSocketFallsBackToPID: when --socket is unreachable but
// --pid-file is provided, daemon stop falls back to SIGTERM on the
// recorded PID. We simulate this by giving a wrong socket path.
func TestDaemonStopSocketFallsBackToPID(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	pidPath := filepath.Join(t.TempDir(), "agentd.pid")

	if err := run([]string{
		"daemon", "start",
		"--socket", sock,
		"--pid-file", pidPath,
		"--lock-file", lockPath,
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		_ = run([]string{"daemon", "stop", "--socket", sock, "--pid-file", pidPath, "--timeout-seconds", "3"})
	})

	// Wrong socket + correct pid-file: stop should still succeed via SIGTERM.
	wrongSock := filepath.Join(t.TempDir(), "wrong.sock")
	if err := run([]string{
		"daemon", "stop",
		"--socket", wrongSock,
		"--pid-file", pidPath,
		"--timeout-seconds", "3",
	}); err != nil {
		t.Fatalf("stop with wrong socket but right pid-file: %v", err)
	}

	// Daemon must be gone.
	if _, err := net.DialTimeout("unix", sock, 200*time.Millisecond); err == nil {
		t.Fatal("daemon still responding after PID-fallback stop")
	}
}

// TestDaemonStopRequiresSocketOrPID: with neither --socket nor
// --pid-file, stop must produce a clear error.
func TestDaemonStopRequiresSocketOrPID(t *testing.T) {
	err := run([]string{"daemon", "stop"})
	if err == nil {
		t.Fatal("expected error when neither --socket nor --pid-file is provided")
	}
}

// TestNeverMatchesAnything is referenced by daemonSpawnHelper to
// suppress all tests in the child process. Each child invocation
// passes `-test.run=TestNeverMatchesAnything --` to the test binary
// before falling through to the helper logic in TestMain.
func TestNeverMatchesAnything(t *testing.T) {}

// ---- helpers used by tests above ----

// daemonForegroundCtxTest re-uses the existing ctx-driven foreground
// path to make sure the background mode doesn't regress it. Sanity
// check only — the real foreground coverage stays in daemon_test.go.
func TestForegroundPathStillWorks(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	go func() {
		errCh <- runDaemonWithContext(ctx, []string{"start", "--foreground", "--socket", sock, "--lock-file", lockPath})
	}()
	if !waitForSocket(sock, 2*time.Second) {
		t.Fatal("foreground daemon did not bind socket")
	}
	cancel()
	select {
	case <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatal("foreground daemon did not exit on ctx cancel")
	}
}
