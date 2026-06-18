package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func applyCompletedResult(db taskdb.TaskDB, taskID string, record taskdb.TaskRecord, now time.Time) (taskdb.TaskDB, error) {
	recordState := record.State.Code()
	if recordState == task.TaskStateCodeValidating ||
		recordState == task.TaskStateCodePatchReady ||
		recordState == task.TaskStateCodeHumanReview {
		return db, nil
	}
	db, err := ensureRunning(db, taskID, now)
	if err != nil {
		return taskdb.TaskDB{}, err
	}
	record, _ = findTask(db, taskID)
	return applyTransition(db, record, task.StateValidating, ir.EventRunReportedDone, "provider reported run done", "run-reported-done", now)
}
