package taskvalidation

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/ir"
	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func maybeApplyValidationTransition(
	updated *taskdb.TaskDB,
	taskBeforeValidation taskdb.TaskRecord,
	req Request,
	providerForRun string,
	result validation.CommandResult,
	now time.Time,
) (*taskdb.TaskTransitionRecord, *taskdb.TaskCommandReceiptRecord, error) {
	if taskBeforeValidation.State != task.StateValidating {
		return nil, nil, nil
	}
	toState, eventType := validationTransitionForResult(result.Result)
	nextDB, transition, receipt, err := taskdb.ApplyGuardedTaskTransition(*updated, taskdb.TaskTransitionInput{
		TaskID:  req.TaskID,
		ToState: toState,
		Event:   eventType,
		Actor:   req.Actor,
		Source:  req.Source,
		Reason:  fmt.Sprintf("validation %s via %s", result.Result, result.ValidationGate),
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   req.CommandID + ":transition",
			Provider:    providerForRun,
			DecisionLLM: req.DecisionLLM,
			ApprovalID:  req.ApprovalID,
		},
	}, now)
	if err != nil {
		return nil, nil, err
	}
	*updated = nextDB
	return &transition, &receipt, nil
}

func validationTransitionForResult(result string) (task.TaskState, ir.EventType) {
	if result == "passed" {
		return task.StatePatchReady, ir.EventValidationPassed
	}
	return task.StateFailed, ir.EventValidationFailed
}
