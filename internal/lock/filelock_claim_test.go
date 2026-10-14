package lock

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFileLockClaimStaleUsesRefreshedAt(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	claim := newFileLockClaim("/tmp/riido.lock", now.Add(-fileLockClaimStaleAfter/2))
	claim.RefreshedAt = now.Add(-fileLockClaimStaleAfter / 3)
	raw, err := encodeFileLockClaim(claim)
	if err != nil {
		t.Fatalf("encodeFileLockClaim: %v", err)
	}

	if fileLockClaimStale(raw, now.Add(-2*fileLockClaimStaleAfter), now) {
		t.Fatal("fresh refreshed_at must prevent active Windows claim recovery")
	}
}

func TestFileLockClaimStaleAllowsOldClaimRecovery(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	claim := newFileLockClaim("/tmp/riido.lock", now.Add(-2*fileLockClaimStaleAfter))
	raw, err := encodeFileLockClaim(claim)
	if err != nil {
		t.Fatalf("encodeFileLockClaim: %v", err)
	}

	if !fileLockClaimStale(raw, now.Add(-2*fileLockClaimStaleAfter), now) {
		t.Fatal("old Windows claim should be recoverable")
	}
}

func TestFileLockClaimStaleFallsBackToModTimeForLegacyClaim(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	raw := []byte("legacy claim")

	if fileLockClaimStale(raw, now.Add(-fileLockClaimStaleAfter/2), now) {
		t.Fatal("fresh legacy claim must not be recovered")
	}
	if !fileLockClaimStale(raw, now.Add(-2*fileLockClaimStaleAfter), now) {
		t.Fatal("old legacy claim should be recoverable from modtime")
	}
}

func TestEncodeFileLockClaimKeepsOwnerMetadata(t *testing.T) {
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)
	raw, err := encodeFileLockClaim(newFileLockClaim("/tmp/riido.lock", now))
	if err != nil {
		t.Fatalf("encodeFileLockClaim: %v", err)
	}
	var got fileLockClaim
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("decode claim: %v", err)
	}
	if got.SchemaVersion != fileLockClaimSchemaVersion || got.PID == 0 || got.Path != "/tmp/riido.lock" ||
		!got.CreatedAt.Equal(now) || !got.RefreshedAt.Equal(now) {
		t.Fatalf("claim metadata = %+v", got)
	}
}
