package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func applyTimeoutResult(db taskdb.TaskDB, record taskdb.TaskRecord, res agentbridge.Result, now time.Time) (taskdb.TaskDB, error) {
	if timeoutCanOriginate(record.State) {
		return applyTransition(db, record, task.StateTimedOut, ir.EventTaskTimedOut, resultReason(res, "provider run timed out"), "timed-out", now)
	}
	return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, "provider timed out before running"), "failed-timeout-before-running", now)
}
