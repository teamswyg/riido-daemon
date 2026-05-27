// Package scheduling owns the C5 Runtime Scheduling domain: "which runtime
// claims which task", heartbeat semantics, and the RuntimeLease meaning
// (who can hold it, when it expires, what fingerprint it binds).
//
// What this package does NOT own:
//   - The flock / DB lease acquisition mechanics → C9 Locking/Lease
//     primitive (internal/lock for local files; remote DB lease is future work).
//   - The Provider port and adapter execution → C4 Provider Runtime.
//
// Dependency direction: scheduling imports capability (read-only).
// capability does NOT import scheduling.
package scheduling

import (
	"time"

	"github.com/teamswyg/riido-contracts/provider/capability"
)

// RuntimeLease is the C5 domain model — "which task is currently claimed by
// which runtime under which capability snapshot".
//
// A lease binds three things:
//
//	(LeaseID, TaskID) — which task this lease is for
//	RuntimeID         — which runtime slot holds it
//	CapabilityFingerprint — the effective capability snapshot the lease
//	                        was issued against
//
// The (RuntimeID, CapabilityFingerprint) pair IS the runtime pin
// between C5 scheduling and the C1 task lifecycle. If either side of the pair
// changes while the lease is alive, the lease MUST be invalidated — a stale
// lease holder must never continue to advance the task.
type RuntimeLease struct {
	LeaseID               string
	TaskID                string
	RuntimeID             capability.RuntimeID
	CapabilityFingerprint capability.CapabilityFingerprint
	ClaimedAt             time.Time
	LeaseUntil            time.Time
	// FencingToken is a monotonic counter per lease lineage. Writers
	// committing under a stale token are rejected by C9 lease primitives.
	FencingToken int64
}

// IsExpired reports whether the lease deadline has passed.
func (l RuntimeLease) IsExpired(now time.Time) bool {
	return now.After(l.LeaseUntil)
}

// IsPinnedTo reports whether the lease still binds the same
// (RuntimeID, CapabilityFingerprint) pair. If either side has changed, the
// lease is no longer the pin — the orchestrator MUST invalidate it.
func (l RuntimeLease) IsPinnedTo(rid capability.RuntimeID, fp capability.CapabilityFingerprint) bool {
	return l.RuntimeID == rid && l.CapabilityFingerprint == fp
}
