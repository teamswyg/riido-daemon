package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/agentbridge"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func applyTerminalResult(db taskdb.TaskDB, taskID string, res agentbridge.Result, now time.Time) (taskdb.TaskDB, error) {
	record, ok := findTask(db, taskID)
	if !ok {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "apply-terminal-result", "task %s not found", taskID)
	}
	if record.State.IsTerminal() {
		return db, nil
	}
	status := terminalResultStatus(res)
	switch status {
	case agentbridge.ResultCompleted:
		return applyCompletedResult(db, taskID, record, now)
	case agentbridge.ResultCancelled:
		return applyTransition(db, record, task.StateCancelled, ir.EventTaskCancelled, resultReason(res, "provider run cancelled"), "cancelled", now)
	case agentbridge.ResultTimeout:
		return applyTimeoutResult(db, record, res, now)
	case agentbridge.ResultBlocked:
		return applyBlockedResult(db, taskID, record, res, now)
	default:
		return applyTransition(db, record, task.StateFailed, ir.EventTaskFailed, resultReason(res, string(status)), "failed:"+string(status), now)
	}
}

func terminalResultStatus(res agentbridge.Result) agentbridge.ResultStatus {
	if res.Status == "" {
		return agentbridge.ResultCompleted
	}
	return res.Status
}
