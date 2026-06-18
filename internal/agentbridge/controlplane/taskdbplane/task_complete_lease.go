package taskdbplane

import "time"

func (p *Plane) releaseCompletedLease(leases RuntimeLeaseRegistry, taskID string, now time.Time) error {
	leases, changed := releaseRuntimeLease(leases, taskID, now)
	if !changed {
		return nil
	}
	return saveRuntimeLeaseRegistry(p.leasePath, p.path, leases, now)
}
