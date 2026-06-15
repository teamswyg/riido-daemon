package taskdbplane

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge/controlplane"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
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

func reconcileExpiredRuntimeLeases(db taskdb.TaskDB, registry RuntimeLeaseRegistry, now time.Time) (taskdb.TaskDB, RuntimeLeaseRegistry, bool, error) {
	changed := false
	for _, lease := range append([]RuntimeLeaseRecord(nil), registry.Leases...) {
		if lease.ReleasedAt != nil || !lease.isExpired(now) {
			continue
		}
		record, ok := findTask(db, lease.TaskID)
		if !ok {
			var released bool
			registry, released = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = changed || released
			continue
		}
		switch record.State.Code() {
		case task.TaskStateCodePreparing, task.TaskStateCodeRunning:
			updated, err := applyExpiredRuntimeHandoff(db, record, lease, now)
			if err != nil {
				return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
			}
			db = updated
			registry, _ = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = true
		case task.TaskStateCodeClaimed:
			updated, err := applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, "runtime lease expired before provider execution", "lease-expired:"+lease.LeaseID+":failed", now)
			if err != nil {
				return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
			}
			db = updated
			registry, _ = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = true
		case task.TaskStateCodeNeedsInput:
			updated, err := applyTransition(db, record, task.StateTimedOut, ir.EventTaskTimedOut, "runtime lease expired while waiting for input", "lease-expired:"+lease.LeaseID+":timed-out", now)
			if err != nil {
				return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
			}
			db = updated
			registry, _ = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = true
		case task.TaskStateCodeUnknown, task.TaskStateCodeQueued, task.TaskStateCodeCreated, task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodePatchReady, task.TaskStateCodeHumanReview, task.TaskStateCodeReworkQueued, task.TaskStateCodeCompleted, task.TaskStateCodeFailed, task.TaskStateCodeCancelled, task.TaskStateCodeTimedOut:
			var released bool
			registry, released = releaseRuntimeLease(registry, lease.TaskID, now)
			changed = changed || released
		}
	}
	return db, registry, changed, nil
}

func applyExpiredRuntimeHandoff(db taskdb.TaskDB, record taskdb.TaskRecord, lease RuntimeLeaseRecord, now time.Time) (taskdb.TaskDB, error) {
	updated, err := applyTransition(db, record, task.StateBlocked, ir.EventBlockerRaised, "runtime lease expired; requeue for another runtime", "lease-expired:"+lease.LeaseID+":blocked", now)
	if err != nil {
		return taskdb.TaskDB{}, err
	}
	blocked, ok := findTask(updated, record.ID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "lease.expire-handoff", "task %s not found after lease expiry block", record.ID)
	}
	return applyTransition(updated, blocked, task.StateQueued, ir.EventBlockerResolvedRequeue, "runtime lease expired; handoff queued", "lease-expired:"+lease.LeaseID+":requeue", now)
}

func requireActiveRuntimeLease(registry RuntimeLeaseRegistry, taskID string, now time.Time, report controlplane.TaskReportContext) (RuntimeLeaseRecord, error) {
	idx := runtimeLeaseIndex(registry.Leases, taskID)
	if idx < 0 {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s has no runtime lease", taskID)
	}
	lease := registry.Leases[idx]
	if !lease.isActive(now) {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime lease is not active", taskID)
	}
	if report.RuntimeLeaseID == "" {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s report missing runtime lease id", taskID)
	}
	if lease.LeaseID != report.RuntimeLeaseID {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime lease id mismatch", taskID)
	}
	if !report.RuntimeFencingTokenSet {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s report missing runtime fencing token", taskID)
	}
	if lease.FencingToken != report.RuntimeFencingToken {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime fencing token mismatch", taskID)
	}
	if report.RuntimeCapabilityFingerprint != "" && lease.CapabilityFingerprint != report.RuntimeCapabilityFingerprint {
		return RuntimeLeaseRecord{}, planeErrorf(ErrTaskDBPlaneLease, "lease.require-active", "task %s runtime capability fingerprint mismatch", taskID)
	}
	return lease, nil
}

func (r RuntimeLeaseRecord) isActive(now time.Time) bool {
	return r.ReleasedAt == nil && !r.LeaseUntil.IsZero() && !now.After(r.LeaseUntil)
}

func (r RuntimeLeaseRecord) isExpired(now time.Time) bool {
	return r.ReleasedAt == nil && !r.LeaseUntil.IsZero() && now.After(r.LeaseUntil)
}

func runtimeLeaseIndex(leases []RuntimeLeaseRecord, taskID string) int {
	for i, lease := range leases {
		if lease.TaskID == taskID {
			return i
		}
	}
	return -1
}

func runtimeLeaseID(taskID string, fencingToken int64) string {
	return fmt.Sprintf("runtime-lease:%s:%d", taskID, fencingToken)
}

func runtimeHasCapabilityFingerprint(rec controlplane.RegisteredRuntime, fingerprint string) bool {
	if strings.TrimSpace(fingerprint) == "" {
		return true
	}
	for key, value := range rec.CapabilityAttributes {
		if strings.HasPrefix(key, "provider.") && strings.HasSuffix(key, ".capability_fingerprint") && value == fingerprint {
			return true
		}
	}
	return false
}

func normalizedTaskIDs(ids []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		out = append(out, trimmed)
	}
	sort.Strings(out)
	return out
}
