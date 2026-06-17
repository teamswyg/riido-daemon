package lock

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReclaimStaleLockClaimRemovesOldClaim(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	lockPath := filepath.Join(t.TempDir(), "riido.lock")
	writeClaim(t, lockPath, now.Add(-2*fileLockClaimStaleAfter))

	if err := reclaimStaleLockClaim(lockPath, now); err != nil {
		t.Fatalf("reclaimStaleLockClaim: %v", err)
	}
	if _, err := os.Stat(lockClaimPath(lockPath)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("claim after reclaim stat error = %v", err)
	}
}

func TestReclaimStaleLockClaimKeepsFreshClaim(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	lockPath := filepath.Join(t.TempDir(), "riido.lock")
	writeClaim(t, lockPath, now.Add(-fileLockClaimStaleAfter/2))

	if err := reclaimStaleLockClaim(lockPath, now); !errors.Is(err, os.ErrExist) {
		t.Fatalf("reclaimStaleLockClaim error = %v, want os.ErrExist", err)
	}
	if _, err := os.Stat(lockClaimPath(lockPath)); err != nil {
		t.Fatalf("fresh claim stat error = %v", err)
	}
}

func TestReclaimStaleLockClaimIgnoresMissingClaim(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "riido.lock")

	if err := reclaimStaleLockClaim(lockPath, time.Now()); err != nil {
		t.Fatalf("reclaimStaleLockClaim missing claim: %v", err)
	}
}

func writeClaim(t *testing.T, lockPath string, refreshedAt time.Time) {
	t.Helper()
	body, err := encodeFileLockClaim(newFileLockClaim(lockPath, refreshedAt))
	if err != nil {
		t.Fatalf("encodeFileLockClaim: %v", err)
	}
	if err := os.WriteFile(lockClaimPath(lockPath), body, 0o644); err != nil {
		t.Fatalf("write claim: %v", err)
	}
}
