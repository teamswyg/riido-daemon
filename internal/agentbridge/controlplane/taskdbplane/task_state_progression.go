package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func ensurePreparing(db taskdb.TaskDB, taskID string, now time.Time) (taskdb.TaskDB, error) {
	record, ok := findTask(db, taskID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-preparing", "task %s not found", taskID)
	}
	switch record.State.Code() {
	case task.TaskStateCodePreparing, task.TaskStateCodeRunning, task.TaskStateCodeNeedsInput,
		task.TaskStateCodeBlocked, task.TaskStateCodeValidating, task.TaskStateCodePatchReady,
		task.TaskStateCodeHumanReview:
		return db, nil
	case task.TaskStateCodeClaimed:
		return applyTransition(db, record, task.StatePreparing, ir.EventWorkdirPreparing, "workspace preparation started", "preparing", now)
	default:
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-preparing", "cannot start task %s from state %s", taskID, record.State)
	}
}

func ensureRunning(db taskdb.TaskDB, taskID string, now time.Time) (taskdb.TaskDB, error) {
	record, ok := findTask(db, taskID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-running", "task %s not found", taskID)
	}
	switch record.State.Code() {
	case task.TaskStateCodeRunning, task.TaskStateCodeNeedsInput, task.TaskStateCodeBlocked,
		task.TaskStateCodeValidating, task.TaskStateCodePatchReady, task.TaskStateCodeHumanReview:
		return db, nil
	case task.TaskStateCodeClaimed:
		nextDB, err := applyTransition(db, record, task.StatePreparing, ir.EventWorkdirPreparing, "workspace preparation started", "preparing", now)
		if err != nil {
			return taskdb.TaskDB{}, err
		}
		record, _ = findTask(nextDB, taskID)
		return transitionPreparingToRunning(nextDB, record, now)
	case task.TaskStateCodePreparing:
		return transitionPreparingToRunning(db, record, now)
	default:
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "ensure-running", "cannot run task %s from state %s", taskID, record.State)
	}
}

func transitionPreparingToRunning(db taskdb.TaskDB, record taskdb.TaskRecord, now time.Time) (taskdb.TaskDB, error) {
	return applyTransition(db, record, task.StateRunning, ir.EventRunStarted, "provider process started", "run-started", now)
}
