package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func applyBlockedResult(db taskdb.TaskDB, taskID string, record taskdb.TaskRecord, res agentbridge.Result, now time.Time) (taskdb.TaskDB, error) {
	recordState := record.State.Code()
	if recordState == task.TaskStateCodeBlocked {
		return db, nil
	}
	if recordState == task.TaskStateCodeClaimed {
		nextDB, err := ensurePreparing(db, taskID, now)
		if err != nil {
			return taskdb.TaskDB{}, err
		}
		db = nextDB
		record, _ = findTask(db, taskID)
		recordState = record.State.Code()
	}
	if recordState == task.TaskStateCodePreparing || recordState == task.TaskStateCodeRunning {
		return applyTransition(db, record, task.StateBlocked, ir.EventBlockerRaised, resultReason(res, "runtime eligibility blocked task"), "blocked", now)
	}
	return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, "runtime eligibility blocked task from invalid state"), "failed:blocked-invalid-state", now)
}
