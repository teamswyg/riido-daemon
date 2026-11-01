//go:build darwin || linux

package securecredential

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestTryLockRejectsContentionAndSeparatesResourceKeys(t *testing.T) {
	a := &Store{Key: "repair-origin-0123456789", LockDir: filepath.Join(t.TempDir(), "locks")}
	unlock, err := a.TryLock(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer unlock()
	b := *a
	if next, err := b.TryLock(t.Context()); next != nil || !errors.Is(err, ErrBusy) {
		t.Fatal("same resource did not return busy", err)
	}
	b.Key = "other-origin-0123456789"
	other, err := b.TryLock(t.Context())
	if err != nil {
		t.Fatal("different resource was blocked", err)
	}
	other()
	unlock()
	unlock()
	next, err := a.TryLock(t.Context())
	if err != nil {
		t.Fatal("released resource remained busy", err)
	}
	next()
}

func TestTryLockPreservesCancellationAndFilesystemGuards(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	s := &Store{Key: "repair-origin-0123456789", LockDir: filepath.Join(t.TempDir(), "absent")}
	if _, err := s.TryLock(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("cancelled call was not cancelled", err)
	}
	if _, err := os.Stat(s.LockDir); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("cancelled call created lock files")
	}
	s.LockDir = filepath.Join(t.TempDir(), "locks")
	if err := os.Mkdir(s.LockDir, 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(s.LockDir, s.Key+".lock")); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TryLock(t.Context()); !errors.Is(err, ErrUnavailable) {
		t.Fatal("symlink accepted", err)
	}
	s.LockDir = t.TempDir()
	if err := os.Chmod(s.LockDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := s.TryLock(t.Context()); !errors.Is(err, ErrUnavailable) {
		t.Fatal("shared directory accepted", err)
	}
}
