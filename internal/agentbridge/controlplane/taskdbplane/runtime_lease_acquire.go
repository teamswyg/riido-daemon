package taskdbplane

import "time"

func acquireRuntimeLease(registry RuntimeLeaseRegistry, taskID, runtimeID, capabilityFingerprint string, now time.Time, ttl time.Duration) (RuntimeLeaseRegistry, RuntimeLeaseRecord, bool) {
	if ttl <= 0 {
		ttl = defaultRuntimeLeaseTTL
	}
	idx := runtimeLeaseIndex(registry.Leases, taskID)
	if idx >= 0 {
		return acquireExistingRuntimeLease(registry, idx, runtimeID, capabilityFingerprint, now, ttl)
	}
	lease := newRuntimeLease(taskID, runtimeID, capabilityFingerprint, now, ttl, 1)
	registry.Leases = append(registry.Leases, lease)
	return registry, lease, true
}

func acquireExistingRuntimeLease(registry RuntimeLeaseRegistry, idx int, runtimeID, capabilityFingerprint string, now time.Time, ttl time.Duration) (RuntimeLeaseRegistry, RuntimeLeaseRecord, bool) {
	existing := registry.Leases[idx]
	if existing.isActive(now) {
		if existing.RuntimeID != runtimeID || existing.CapabilityFingerprint != capabilityFingerprint {
			return registry, RuntimeLeaseRecord{}, false
		}
		existing.LeaseUntil = now.Add(ttl)
		registry.Leases[idx] = existing
		return registry, existing, true
	}
	lease := newRuntimeLease(existing.TaskID, runtimeID, capabilityFingerprint, now, ttl, existing.FencingToken+1)
	registry.Leases[idx] = lease
	return registry, lease, true
}

func newRuntimeLease(taskID, runtimeID, capabilityFingerprint string, now time.Time, ttl time.Duration, fencingToken int64) RuntimeLeaseRecord {
	return RuntimeLeaseRecord{
		LeaseID:               runtimeLeaseID(taskID, fencingToken),
		TaskID:                taskID,
		RuntimeID:             runtimeID,
		CapabilityFingerprint: capabilityFingerprint,
		ClaimedAt:             now,
		LeaseUntil:            now.Add(ttl),
		FencingToken:          fencingToken,
	}
}
