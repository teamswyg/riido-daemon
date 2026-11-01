//go:build darwin || linux

package securecredential

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestTryLockProcessHelper(t *testing.T) {
	dir := os.Getenv("RIIDO_TRY_LOCK_TEST_DIR")
	if dir == "" {
		t.Skip("only run by parent process test")
	}
	s := &Store{Key: "repair-origin-0123456789", LockDir: dir}
	unlock, err := s.TryLock(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	fmt.Println("locked")
	<-t.Context().Done()
}

func TestTryLockRejectsOtherProcessAndReleasesAfterKill(t *testing.T) {
	s := &Store{Key: "repair-origin-0123456789", LockDir: filepath.Join(t.TempDir(), "locks")}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestTryLockProcessHelper$", "-test.timeout=15s")
	cmd.Env = []string{"RIIDO_TRY_LOCK_TEST_DIR=" + s.LockDir, "PATH=/usr/bin:/bin"}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if cmd.ProcessState == nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	}()
	ready := make(chan string, 1)
	go func() { line, _ := bufio.NewReader(out).ReadString('\n'); ready <- line }()
	select {
	case line := <-ready:
		if line != "locked\n" {
			t.Fatal("child failed to acquire lock")
		}
	case <-ctx.Done():
		t.Fatal("child did not become ready")
	}
	if unlock, err := s.TryLock(ctx); unlock != nil || !errors.Is(err, ErrBusy) {
		t.Fatal("live process lock was not busy", err)
	}
	if err = cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	var exited *exec.ExitError
	if err = cmd.Wait(); !errors.As(err, &exited) {
		t.Fatal("child did not exit after kill", err)
	}
	unlock, err := s.TryLock(ctx)
	if err != nil {
		t.Fatal("terminated process retained lock", err)
	}
	unlock()
}
