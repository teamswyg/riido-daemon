package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func reconcileExpiredRuntimeLeases(db taskdb.TaskDB, registry RuntimeLeaseRegistry, now time.Time) (taskdb.TaskDB, RuntimeLeaseRegistry, bool, error) {
	changed := false
	for _, lease := range append([]RuntimeLeaseRecord(nil), registry.Leases...) {
		if lease.ReleasedAt != nil || !lease.isExpired(now) {
			continue
		}
		nextDB, nextRegistry, released, err := reconcileExpiredRuntimeLease(db, registry, lease, now)
		if err != nil {
			return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
		}
		db = nextDB
		registry = nextRegistry
		changed = changed || released
	}
	return db, registry, changed, nil
}

func reconcileExpiredRuntimeLease(db taskdb.TaskDB, registry RuntimeLeaseRegistry, lease RuntimeLeaseRecord, now time.Time) (taskdb.TaskDB, RuntimeLeaseRegistry, bool, error) {
	record, ok := findTask(db, lease.TaskID)
	if !ok {
		next, released := releaseRuntimeLease(registry, lease.TaskID, now)
		return db, next, released, nil
	}
	nextDB, err := applyExpiredRuntimeLeaseTransition(db, record, lease, now)
	if err != nil {
		return taskdb.TaskDB{}, RuntimeLeaseRegistry{}, false, err
	}
	nextRegistry, released := releaseRuntimeLease(registry, lease.TaskID, now)
	return nextDB, nextRegistry, released, nil
}
