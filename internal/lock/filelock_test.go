package lock

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestTryAcquireFileReturnsErrLockedWhileHeld(t *testing.T) {
	path := filepath.Join(t.TempDir(), "daemon.lock")
	first, err := TryAcquireFile(path)
	if err != nil {
		t.Fatalf("TryAcquireFile first: %v", err)
	}
	if _, err := TryAcquireFile(path); !errors.Is(err, ErrLocked) {
		t.Fatalf("second TryAcquireFile = %v, want ErrLocked", err)
	}
	if err := first.Release(); err != nil {
		t.Fatalf("Release first: %v", err)
	}
	second, err := TryAcquireFile(path)
	if err != nil {
		t.Fatalf("TryAcquireFile after release: %v", err)
	}
	if err := second.Release(); err != nil {
		t.Fatalf("Release second: %v", err)
	}
}

func TestTryAcquireFileEmptyPath(t *testing.T) {
	if _, err := TryAcquireFile("  "); err == nil {
		t.Fatal("TryAcquireFile with empty path should error")
	}
}

func TestRemoveStaleLockUnixNoOp(t *testing.T) {
	// On Unix the flock is released on process death, so RemoveStaleLock is a
	// no-op and must not error even when the artifact is absent.
	if err := RemoveStaleLock(filepath.Join(t.TempDir(), "missing.lock")); err != nil {
		t.Fatalf("RemoveStaleLock: %v", err)
	}
}

func TestWithFileSerializesExclusiveLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "task-db.lock")
	first, err := AcquireFile(context.Background(), path)
	if err != nil {
		t.Fatalf("AcquireFile first: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := AcquireFile(ctx, path); err == nil {
		t.Fatal("second acquire should time out while first lock is held")
	}
	if err := first.Release(); err != nil {
		t.Fatalf("Release first: %v", err)
	}
	if err := WithFile(context.Background(), path, func() error { return nil }); err != nil {
		t.Fatalf("WithFile after release: %v", err)
	}
}
