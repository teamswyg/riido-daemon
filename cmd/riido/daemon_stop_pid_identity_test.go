package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDaemonStopRejectsNonDaemonPidFile(t *testing.T) {
	pidPath := filepath.Join(t.TempDir(), "agentd.pid")
	shim, waitDone := startNonDaemonPIDShim(t)

	if err := os.WriteFile(pidPath, []byte(strconv.Itoa(shim.Process.Pid)), 0o644); err != nil {
		t.Fatalf("write pid: %v", err)
	}

	err := run([]string{"daemon", "stop", "--pid-file", pidPath, "--timeout-seconds", "1"})
	if err == nil || !strings.Contains(err.Error(), "pid identity") {
		t.Fatalf("expected non-daemon pid rejection, got %v", err)
	}

	select {
	case <-waitDone:
		t.Fatal("non-daemon pid was signaled")
	case <-time.After(150 * time.Millisecond):
	}
}

func startNonDaemonPIDShim(t *testing.T) (*exec.Cmd, <-chan error) {
	t.Helper()
	shim := exec.Command("/bin/sleep", "30")
	if err := shim.Start(); err != nil {
		t.Fatalf("shim: %v", err)
	}
	waitDone := make(chan error, 1)
	go func() { waitDone <- shim.Wait() }()
	t.Cleanup(func() { _ = shim.Process.Kill() })
	return shim, waitDone
}
