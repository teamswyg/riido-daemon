package taskdb

import (
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
)

func ApplyTaskTransition(existing TaskDB, taskID string, to task.TaskState, event ir.EventType, actor, source, reason string, now time.Time) (TaskDB, TaskTransitionRecord, error) {
	updated, transition, _, err := ApplyGuardedTaskTransition(existing, TaskTransitionInput{
		TaskID:  taskID,
		ToState: to,
		Event:   event,
		Actor:   actor,
		Source:  source,
		Reason:  reason,
		Guard: TaskMutationGuardInput{
			ApprovalID: "approval.riido.legacy",
		},
	}, now)
	return updated, transition, err
}

func AddTaskEvidence(existing TaskDB, input TaskEvidenceInput, now time.Time) (TaskDB, TaskEvidenceRecord, error) {
	updated, evidence, _, err := AddGuardedTaskEvidence(existing, input, now)
	return updated, evidence, err
}

func ParseTaskState(value string) (task.TaskState, error) {
	code := task.ParseTaskStateCode(value)
	if !code.IsKnown() {
		return "", taskDBErrorf(ErrTaskDBState, "parse-state", "unknown task state: %s", value)
	}
	return code.TaskState(), nil
}
