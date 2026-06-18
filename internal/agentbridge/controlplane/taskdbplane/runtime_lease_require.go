package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
)

func requireActiveRuntimeLease(registry RuntimeLeaseRegistry, taskID string, now time.Time, report controlplane.TaskReportContext) (RuntimeLeaseRecord, error) {
	idx := runtimeLeaseIndex(registry.Leases, taskID)
	if idx < 0 {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s has no runtime lease", taskID)
	}
	lease := registry.Leases[idx]
	if !lease.isActive(now) {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime lease is not active", taskID)
	}
	if err := requireLeaseReportIdentity(taskID, lease, report); err != nil {
		return RuntimeLeaseRecord{}, err
	}
	return lease, nil
}

func requireLeaseReportIdentity(taskID string, lease RuntimeLeaseRecord, report controlplane.TaskReportContext) error {
	if report.RuntimeLeaseID == "" {
		return planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s report missing runtime lease id", taskID)
	}
	if lease.LeaseID != report.RuntimeLeaseID {
		return planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime lease id mismatch", taskID)
	}
	if !report.RuntimeFencingTokenSet {
		return planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s report missing runtime fencing token", taskID)
	}
	if lease.FencingToken != report.RuntimeFencingToken {
		return planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime fencing token mismatch", taskID)
	}
	if report.RuntimeCapabilityFingerprint != "" && lease.CapabilityFingerprint != report.RuntimeCapabilityFingerprint {
		return planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime capability fingerprint mismatch", taskID)
	}
	return nil
}
