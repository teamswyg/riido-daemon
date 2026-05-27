package scheduling

import (
	"testing"
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

func TestLeaseExpiry(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)
	l := RuntimeLease{LeaseUntil: now.Add(5 * time.Minute)}
	if l.IsExpired(now) {
		t.Error("lease must not be expired before its deadline")
	}
	if !l.IsExpired(now.Add(10 * time.Minute)) {
		t.Error("lease must be expired past its deadline")
	}
}

func TestLeasePinningInvariant(t *testing.T) {
	rid := capability.RuntimeID("rt-1")
	fp := capability.CapabilityFingerprint("fp-abc")
	l := RuntimeLease{RuntimeID: rid, CapabilityFingerprint: fp}

	if !l.IsPinnedTo(rid, fp) {
		t.Error("lease must remain pinned to its original (RuntimeID, CapabilityFingerprint) pair")
	}
	if l.IsPinnedTo(rid, capability.CapabilityFingerprint("fp-xyz")) {
		t.Error("lease must NOT match a different CapabilityFingerprint (runtime pin invariant)")
	}
	if l.IsPinnedTo(capability.RuntimeID("rt-2"), fp) {
		t.Error("lease must NOT match a different RuntimeID (runtime pin invariant)")
	}
}
