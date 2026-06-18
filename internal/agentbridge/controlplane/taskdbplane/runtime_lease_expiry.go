package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func applyExpiredRuntimeLeaseTransition(db taskdb.TaskDB, record taskdb.TaskRecord, lease RuntimeLeaseRecord, now time.Time) (taskdb.TaskDB, error) {
	switch record.State.Code() {
	case task.TaskStateCodePreparing, task.TaskStateCodeRunning:
		return applyExpiredRuntimeHandoff(db, record, lease, now)
	case task.TaskStateCodeClaimed:
		return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, "runtime lease expired before provider execution", "lease-expired:"+lease.LeaseID+":failed", now)
	case task.TaskStateCodeNeedsInput:
		return applyTransition(db, record, task.StateTimedOut, ir.EventTaskTimedOut, "runtime lease expired while waiting for input", "lease-expired:"+lease.LeaseID+":timed-out", now)
	case task.TaskStateCodeUnknown, task.TaskStateCodeQueued, task.TaskStateCodeCreated, task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodePatchReady, task.TaskStateCodeHumanReview, task.TaskStateCodeReworkQueued, task.TaskStateCodeCompleted, task.TaskStateCodeFailed, task.TaskStateCodeCancelled, task.TaskStateCodeTimedOut:
		return db, nil
	default:
		return db, nil
	}
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
