//go:build !windows

package childreg

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

// TestReapOrphansKillsProcessGroup spawns a real child in its own process group
// (as the daemon does via Setpgid), records its pgid, and verifies ReapOrphans
// kills the group and clears the file.
func TestReapOrphansKillsProcessGroup(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	pid := cmd.Process.Pid
	// Reap the process at test end no matter what so we never leak it.
	defer func() {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
		_, _ = cmd.Process.Wait()
	}()

	path := filepath.Join(t.TempDir(), "daemon-children.pids")
	if err := os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0o644); err != nil {
		t.Fatalf("write registry: %v", err)
	}

	reaped, err := ReapOrphans(path)
	if err != nil {
		t.Fatalf("ReapOrphans: %v", err)
	}
	if reaped != 1 {
		t.Fatalf("reaped = %d, want 1", reaped)
	}

	// The killed child must be reaped by Wait, then no longer be signalable.
	_, _ = cmd.Process.Wait()
	if err := syscall.Kill(-pid, 0); err == nil {
		t.Fatal("process group should be dead after reap")
	}

	// The registry file is reset, so a second reap finds nothing.
	if data, _ := os.ReadFile(path); len(strings.Fields(string(data))) != 0 {
		t.Fatalf("registry should be cleared after reap, got %q", string(data))
	}
	if reaped2, _ := ReapOrphans(path); reaped2 != 0 {
		t.Fatalf("second reap = %d, want 0", reaped2)
	}
}

// A live registry whose process exits cleanly leaves nothing to reap.
func TestOnExitRemovesBeforeReap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "daemon-children.pids")
	r := New(path)
	r.OnSpawn(999999) // some pid
	r.OnExit(999999)
	if reaped, _ := ReapOrphans(path); reaped != 0 {
		t.Fatalf("reaped = %d, want 0 (exited child should be removed)", reaped)
	}
}
