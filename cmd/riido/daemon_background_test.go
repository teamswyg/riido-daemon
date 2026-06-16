package main

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestMain detects the helper-process invocation pattern used by the
// background-mode tests. When RIIDO_DAEMON_HELPER=1 is set, the test
// binary acts as the riido CLI and runs `run(args)`. Otherwise it
// installs a daemonSpawnHelper override so daemon-background tests can
// spawn THIS test binary as the daemon child.
func TestMain(m *testing.M) {
	if os.Getenv("RIIDO_DAEMON_HELPER") == "1" {
		// Find the "--" sentinel; everything after it is the riido args.
		var args []string
		for i, a := range os.Args {
			if a == "--" {
				args = os.Args[i+1:]
				break
			}
		}
		if err := run(args); err != nil {
			_, _ = os.Stderr.WriteString("helper: " + err.Error() + "\n")
			os.Exit(1)
		}
		os.Exit(0)
	}
	// Test mode: override the daemon spawner so tests can fork the test
	// binary instead of the production riido executable.
	daemonSpawnHelper = func(args []string) (*exec.Cmd, error) {
		spec := []string{"-test.run=TestNeverMatchesAnything"}
		spec = append(spec, "--")
		spec = append(spec, args...)
		cmd := exec.Command(os.Args[0], spec...)
		cmd.Env = append(os.Environ(), "RIIDO_DAEMON_HELPER=1")
		return cmd, nil
	}
	os.Exit(m.Run())
}

// waitForSocket polls the Unix socket path until it accepts a
// connection or deadline expires.
func waitForSocket(sock string, deadline time.Duration) bool {
	end := time.Now().Add(deadline)
	for time.Now().Before(end) {
		conn, err := net.DialTimeout("unix", sock, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

// TestDaemonBackgroundStartLaunchesChildAndReturns: invoke `daemon
// start` WITHOUT --foreground; the call must return promptly with the
// daemon already running and responding to status.
func TestDaemonBackgroundStartLaunchesChildAndReturns(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	pidPath := filepath.Join(t.TempDir(), "agentd.pid")
	logPath := filepath.Join(t.TempDir(), "agentd.log")

	started := time.Now()
	err := run([]string{
		"daemon", "start",
		"--socket", sock,
		"--pid-file", pidPath,
		"--log-file", logPath,
		"--lock-file", lockPath,
	})
	if err != nil {
		t.Fatalf("background start: %v", err)
	}
	if elapsed := time.Since(started); elapsed > 8*time.Second {
		t.Fatalf("background start blocked too long: %v", elapsed)
	}

	t.Cleanup(func() {
		_ = run([]string{"daemon", "stop", "--socket", sock, "--pid-file", pidPath, "--timeout-seconds", "3"})
	})

	// Daemon must respond to status immediately after start returns.
	out, err := runCapturingStdout(t, func() error {
		return run([]string{"daemon", "status", "--socket", sock})
	})
	if err != nil {
		t.Fatalf("status: %v\n%s", err, out)
	}
	var status struct {
		Health string `json:"health"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		t.Fatalf("parse status: %v\n%s", err, out)
	}
	if status.Health != "ok" {
		t.Fatalf("health: %q\nfull: %s", status.Health, out)
	}

	// PID file must contain the child's PID.
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid pid: %q (err=%v)", pidBytes, err)
	}
	if pid == os.Getpid() {
		t.Fatalf("pid file holds parent's pid (%d) — daemon did not detach", pid)
	}
}

func TestDaemonBackgroundStartRejectsExistingDaemon(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	pidPath := filepath.Join(t.TempDir(), "agentd.pid")
	logPath := filepath.Join(t.TempDir(), "agentd.log")

	startArgs := []string{
		"daemon", "start",
		"--socket", sock,
		"--pid-file", pidPath,
		"--log-file", logPath,
		"--lock-file", lockPath,
	}
	if err := run(startArgs); err != nil {
		t.Fatalf("first background start: %v", err)
	}
	t.Cleanup(func() {
		_ = run([]string{"daemon", "stop", "--socket", sock, "--pid-file", pidPath, "--timeout-seconds", "3"})
	})

	firstPID := readDaemonPIDFile(t, pidPath)
	err := run(startArgs)
	if err == nil {
		t.Fatal("expected duplicate background start to fail while singleton daemon is running")
	}
	if secondPID := readDaemonPIDFile(t, pidPath); secondPID != firstPID {
		t.Fatalf("pid file changed after duplicate start: first=%d second=%d", firstPID, secondPID)
	}
}

// TestDaemonBackgroundLogFileReceivesChildOutput: the child's structured
// log should land in --log-file passed at start.
func TestDaemonBackgroundLogFileReceivesChildOutput(t *testing.T) {
	sock := daemonSocketPath(t)
	lockPath := daemonLockPath(t)
	pidPath := filepath.Join(t.TempDir(), "agentd.pid")
	logPath := filepath.Join(t.TempDir(), "agentd.log")

	if err := run([]string{
		"daemon", "start",
		"--socket", sock,
		"--pid-file", pidPath,
		"--log-file", logPath,
		"--lock-file", lockPath,
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		_ = run([]string{"daemon", "stop", "--socket", sock, "--pid-file", pidPath, "--timeout-seconds", "3"})
	})

	// Trigger one status call to ensure the daemon logs something.
	_, _ = runCapturingStdout(t, func() error {
		return run([]string{"daemon", "status", "--socket", sock})
	})

	// Wait briefly for the daemon to flush.
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

func readDaemonPIDFile(t *testing.T, path string) int {
	t.Helper()
	pidBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil || pid <= 0 {
		t.Fatalf("invalid pid: %q (err=%v)", pidBytes, err)
	}
	return pid
}
