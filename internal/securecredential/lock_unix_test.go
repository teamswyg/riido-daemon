//go:build darwin || linux

package securecredential

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLockSerializesStoresAndHonorsCancellation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "locks")
	a := &Store{Key: "account-0123456789", LockDir: dir}
	b := &Store{Key: "account-0123456789", LockDir: dir}
	unlock, err := a.Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	_, err = b.Lock(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("lock error=%v", err)
	}
	unlock()
	unlock()
	unlock, err = b.Lock(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	unlock()
}

func TestLockRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(dir, "account-0123456789.lock")); err != nil {
		t.Fatal(err)
	}
	if _, err := (&Store{Key: "account-0123456789", LockDir: dir}).Lock(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error=%v", err)
	}
}

func TestLockRejectsSharedDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := (&Store{Key: "account-0123456789", LockDir: dir}).Lock(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error=%v", err)
	}
}
