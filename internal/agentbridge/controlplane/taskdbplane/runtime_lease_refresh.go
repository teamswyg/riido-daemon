package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func refreshRuntimeLeases(registry RuntimeLeaseRegistry, rec controlplane.RegisteredRuntime, taskIDs []string, now time.Time, ttl time.Duration) (RuntimeLeaseRegistry, bool) {
	if ttl <= 0 {
		ttl = defaultRuntimeLeaseTTL
	}
	changed := false
	for _, taskID := range normalizedTaskIDs(taskIDs) {
		idx := runtimeLeaseIndex(registry.Leases, taskID)
		if idx < 0 {
			continue
		}
		lease := registry.Leases[idx]
		if !lease.isActive(now) || lease.RuntimeID != rec.RuntimeID {
			continue
		}
		if !runtimeHasCapabilityFingerprint(rec, lease.CapabilityFingerprint) {
			continue
		}
		lease.LeaseUntil = now.Add(ttl)
		registry.Leases[idx] = lease
		changed = true
	}
	return registry, changed
}
