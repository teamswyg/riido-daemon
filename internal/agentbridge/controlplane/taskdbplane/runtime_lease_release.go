package taskdbplane

import "time"

func releaseRuntimeLease(registry RuntimeLeaseRegistry, taskID string, now time.Time) (RuntimeLeaseRegistry, bool) {
	idx := runtimeLeaseIndex(registry.Leases, taskID)
	if idx < 0 {
		return registry, false
	}
	if registry.Leases[idx].ReleasedAt != nil {
		return registry, false
	}
	releasedAt := now.UTC()
	registry.Leases[idx].ReleasedAt = &releasedAt
	return registry, true
}
