//go:build !windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// deadPID starts and reaps a trivial process so its PID refers to a process that
// has certainly exited (PID reuse within a single test run is vanishingly rare).
func deadPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("/bin/sh", "-c", "exit 0")
	if err := cmd.Run(); err != nil {
		t.Fatalf("run throwaway process: %v", err)
	}
	return cmd.Process.Pid
}

func TestDaemonPIDProbablyAlive(t *testing.T) {
	if !daemonPIDProbablyAlive(os.Getpid()) {
		t.Fatal("current process should be reported alive")
	}
	if daemonPIDProbablyAlive(deadPID(t)) {
		t.Fatal("an exited process should be reported not alive")
	}
	if daemonPIDProbablyAlive(0) {
		t.Fatal("pid 0 should be reported not alive")
	}
}

// On Unix the flock is held by the live first holder, so even a pidfile naming a
// dead owner must NOT let a second caller steal a genuinely-held lock.
func TestAcquireDaemonSingletonDoesNotStealHeldLockOnStalePID(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "daemon.lock")
	pidPath := filepath.Join(dir, "daemon.pid")

	held, alreadyRunning, err := acquireDaemonSingleton(lockPath, pidPath)
	if err != nil || alreadyRunning || held == nil {
		t.Fatalf("first acquire = (held=%v, alreadyRunning=%v, err=%v)", held != nil, alreadyRunning, err)
	}
	defer held.Release()

	// Record a PID that has already exited.
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(deadPID(t))), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}

	second, alreadyRunning, err := acquireDaemonSingleton(lockPath, pidPath)
	if err != nil {
		t.Fatalf("second acquire: %v", err)
	}
	if !alreadyRunning || second != nil {
		t.Fatalf("second acquire = (held=%v, alreadyRunning=%v), want alreadyRunning (held flock not stolen)", second != nil, alreadyRunning)
	}
}
