package lock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAcquireFileReturnsCanceledWhenBusyContextIsDone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "riido.lock")
	first, err := AcquireFile(context.Background(), path)
	if err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	defer first.Release()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = AcquireFile(ctx, path)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("busy canceled acquire error = %v, want context canceled", err)
	}
}

func TestReleaseStopsMaintenanceAndClearsState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "riido.lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if err := tryLockFile(file, path); err != nil {
		t.Fatalf("lock test file: %v", err)
	}
	stopped := false
	lock := &FileLock{file: file, path: path, stopMaintenance: func() { stopped = true }}

	if err := lock.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if !stopped {
		t.Fatal("release did not stop maintenance")
	}
	if lock.file != nil || lock.path != "" || lock.stopMaintenance != nil {
		t.Fatalf("release did not clear state: %+v", lock)
	}
}

func TestFileLockClaimStaleRejectsFutureOrUnknownAge(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	if fileLockClaimStale([]byte("legacy"), time.Time{}, now) {
		t.Fatal("unknown legacy claim age must not be reclaimed")
	}
	claim := newFileLockClaim("/tmp/riido.lock", now)
	claim.RefreshedAt = now.Add(time.Second)
	raw, err := encodeFileLockClaim(claim)
	if err != nil {
		t.Fatal(err)
	}
	if fileLockClaimStale(raw, now.Add(-2*fileLockClaimStaleAfter), now) {
		t.Fatal("future refreshed_at must not be reclaimed")
	}
}
