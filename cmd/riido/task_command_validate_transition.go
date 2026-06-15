package main

import (
	"fmt"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func maybeApplyValidationTransition(
	updated *taskdb.TaskDB,
	taskBeforeValidation taskdb.TaskRecord,
	options taskValidateCLI,
	providerForRun string,
	result validation.CommandResult,
	now time.Time,
) (*taskdb.TaskTransitionRecord, *taskdb.TaskCommandReceiptRecord, error) {
	if taskBeforeValidation.State != task.StateValidating {
		return nil, nil, nil
	}
	toState, eventType := validationTransitionForResult(result.Result)
	nextDB, transition, receipt, err := taskdb.ApplyGuardedTaskTransition(*updated, taskdb.TaskTransitionInput{
		TaskID:  options.taskID,
		ToState: toState,
		Event:   eventType,
		Actor:   options.actor,
		Source:  options.source,
		Reason:  fmt.Sprintf("validation %s via %s", result.Result, result.ValidationGate),
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   options.commandID + ":transition",
			Provider:    providerForRun,
			DecisionLLM: options.decisionLLM,
			ApprovalID:  options.approvalID,
		},
	}, now)
	if err != nil {
		return nil, nil, err
	}
	*updated = nextDB
	return &transition, &receipt, nil
}
