package taskdbplane

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func applyTransition(db taskdb.TaskDB, record taskdb.TaskRecord, to task.TaskState, event ir.EventType, reason, commandSuffix string, now time.Time) (taskdb.TaskDB, error) {
	approvalID := approvalIDForTask(db, record.ID)
	if requiresApproval(db, record) && approvalID == "" {
		return taskdb.TaskDB{}, planeErrorf(ErrTaskDBPlaneTaskState, "apply-transition", "task %s requires approval_id before %s", record.ID, event)
	}
	updated, _, _, err := taskdb.ApplyGuardedTaskTransition(db, taskdb.TaskTransitionInput{
		TaskID:  record.ID,
		ToState: to,
		Event:   event,
		Actor:   defaultActor,
		Source:  sourceName,
		Reason:  reason,
		Guard:   guardFor(db, record, commandSuffix, approvalID),
	}, now)
	if err != nil {
		return taskdb.TaskDB{}, err
	}
	return updated, nil
}

func guardFor(db taskdb.TaskDB, record taskdb.TaskRecord, suffix, approvalID string) taskdb.TaskMutationGuardInput {
	return taskdb.TaskMutationGuardInput{
		CommandID:   commandIDPrefix + record.ID + ":" + suffix,
		Provider:    providerFor(db, record),
		DecisionLLM: decisionLLMFor(db, record),
		ApprovalID:  approvalID,
	}
}
