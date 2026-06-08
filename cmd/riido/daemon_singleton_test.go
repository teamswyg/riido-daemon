package main

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestAcquireDaemonSingletonSecondCallReportsAlreadyRunning(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "daemon.lock")
	pidPath := filepath.Join(dir, "daemon.pid")

	held, alreadyRunning, err := acquireDaemonSingleton(lockPath, pidPath)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if alreadyRunning || held == nil {
		t.Fatalf("first acquire = (held=%v, alreadyRunning=%v), want a held lock", held != nil, alreadyRunning)
	}
	// Simulate the holder recording its live PID, as runDaemonStartForeground does.
	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}

	second, alreadyRunning, err := acquireDaemonSingleton(lockPath, pidPath)
	if err != nil {
		t.Fatalf("second acquire: %v", err)
	}
	if !alreadyRunning || second != nil {
		t.Fatalf("second acquire = (held=%v, alreadyRunning=%v), want alreadyRunning with no lock", second != nil, alreadyRunning)
	}

	if err := held.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	third, alreadyRunning, err := acquireDaemonSingleton(lockPath, pidPath)
	if err != nil {
		t.Fatalf("third acquire after release: %v", err)
	}
	if alreadyRunning || third == nil {
		t.Fatalf("third acquire = (held=%v, alreadyRunning=%v), want a fresh lock", third != nil, alreadyRunning)
	}
	_ = third.Release()
}

func TestDaemonSocketServing(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "agentd.sock")
	if daemonSocketServing(socketPath) {
		t.Fatal("no listener yet: daemonSocketServing should be false")
	}
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	if !daemonSocketServing(socketPath) {
		t.Fatal("with a live listener: daemonSocketServing should be true")
	}
}
