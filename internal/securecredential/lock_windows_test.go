//go:build windows

package securecredential

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func windowsLockTestStore(t *testing.T) *Store {
	t.Helper()
	return &Store{Key: "account-0123456789", LockDir: filepath.Join(t.TempDir(), "locks")}
}

func TestWindowsCredentialLockContentionScopesAndRelease(t *testing.T) {
	s := windowsLockTestStore(t)
	unlock, err := s.Lock(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	copyStore := *s
	if next, err := copyStore.TryLock(t.Context()); next != nil || !errors.Is(err, ErrBusy) {
		t.Fatal("same scope did not reject contention", err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if next, err := copyStore.Lock(ctx); next != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("waiting lock did not respect cancellation", err)
	}
	copyStore.Key = "other-account-0123456789"
	other, err := copyStore.TryLock(t.Context())
	if err != nil {
		t.Fatal("independent scope was blocked", err)
	}
	other()
	unlock()
	unlock()
	next, err := s.TryLock(t.Context())
	if err != nil {
		t.Fatal("released scope remained busy", err)
	}
	next()
	if data, err := os.ReadFile(filepath.Join(s.LockDir, s.Key+".lock")); err != nil || len(data) != 0 {
		t.Fatal("lock file must remain empty and persistent", err)
	}
}

func TestWindowsCredentialLockCancellationAndInvalidPaths(t *testing.T) {
	s := windowsLockTestStore(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	for _, lock := range []func(context.Context) (func(), error){s.Lock, s.TryLock} {
		if unlock, err := lock(ctx); unlock != nil || !errors.Is(err, context.Canceled) {
			t.Fatal("cancelled call was not cancelled", err)
		}
	}
	if _, err := os.Stat(s.LockDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled call created the directory", err)
	}
	for _, path := range []string{"", "relative", `C:relative`, `\\server\share\locks`, `\\?\C:\locks`, `\\.\pipe\locks`, `C:\`, `C:\base\..\locks`, `C:\base\locks.`, `C:\base\locks `, `C:\base\locks:stream`, `C:\base\NUL`} {
		t.Run(path, func(t *testing.T) {
			if _, _, err := windowsLockPath(path); !errors.Is(err, ErrUnavailable) {
				t.Fatal("unsafe path accepted")
			}
		})
	}
	s.Key = "../outside"
	if unlock, err := s.TryLock(t.Context()); unlock != nil || !errors.Is(err, ErrUnavailable) {
		t.Fatal("unsafe key accepted", err)
	}
}

func TestWindowsCredentialLockConcurrentInitialCreation(t *testing.T) {
	s := windowsLockTestStore(t)
	start := make(chan struct{})
	var active atomic.Int32
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			<-start
			unlock, err := s.Lock(t.Context())
			if err != nil {
				t.Error("initial creation failed", err)
				return
			}
			if active.Add(1) != 1 {
				t.Error("simultaneous owners")
			}
			time.Sleep(5 * time.Millisecond)
			active.Add(-1)
			unlock()
		})
	}
	close(start)
	group.Wait()
}

func TestWindowsCredentialLockDoesNotEnableProtectedStore(t *testing.T) {
	s := windowsLockTestStore(t)
	unlock, err := s.TryLock(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	if err := verifyHelper(t.Context(), filepath.Join(t.TempDir(), "helper.exe")); !errors.Is(err, ErrUnavailable) {
		t.Fatal("lock success enabled an unverified helper", err)
	}
}
