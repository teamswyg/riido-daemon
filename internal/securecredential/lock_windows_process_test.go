//go:build windows

package securecredential

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestWindowsCredentialLockProcessHelper(t *testing.T) {
	dir := os.Getenv("RIIDO_WINDOWS_LOCK_TEST_DIR")
	if dir == "" {
		t.Skip("only run as a lock test subprocess")
	}
	s := &Store{Key: "account-0123456789", LockDir: dir}
	unlock, err := s.TryLock(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	fmt.Println("locked")
	<-t.Context().Done()
}

func TestWindowsCredentialLockProcessKillReleases(t *testing.T) {
	s := windowsLockTestStore(t)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWindowsCredentialLockProcessHelper$", "-test.timeout=20s")
	// Keep Windows runtime essentials without inheriting application credentials.
	for _, name := range []string{"SystemRoot", "SystemDrive", "TEMP", "TMP", "PATH"} {
		if value, ok := os.LookupEnv(name); ok {
			cmd.Env = append(cmd.Env, name+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "RIIDO_WINDOWS_LOCK_TEST_DIR="+s.LockDir)
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
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
			t.Fatal("child did not acquire its test lock")
		}
	case <-ctx.Done():
		t.Fatal("child startup timed out")
	}
	if unlock, err := s.TryLock(ctx); unlock != nil || !errors.Is(err, ErrBusy) {
		t.Fatal("live child lock was not busy", err)
	}
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	var exit *exec.ExitError
	if err := cmd.Wait(); !errors.As(err, &exit) {
		t.Fatal("child did not terminate after kill", err)
	}
	// Windows documents that OS lock cleanup can lag process termination.
	for {
		unlock, err := s.TryLock(ctx)
		if err == nil {
			unlock()
			return
		}
		if !errors.Is(err, ErrBusy) {
			t.Fatal("terminated process lock could not be reacquired", err)
		}
		select {
		case <-ctx.Done():
			t.Fatal("OS did not release the terminated process lock")
		case <-time.After(20 * time.Millisecond):
		}
	}
}
