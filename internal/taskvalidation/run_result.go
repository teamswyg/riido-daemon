package taskvalidation

import (
	"fmt"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func buildValidationResult(
	updated taskdb.TaskDB,
	req Request,
	commandResult validation.CommandResult,
	evidence taskdb.TaskEvidenceRecord,
	receipt taskdb.TaskCommandReceiptRecord,
	transition *taskdb.TaskTransitionRecord,
	transitionReceipt *taskdb.TaskCommandReceiptRecord,
	providerForRun string,
) (Result, error) {
	record, ok := FindTask(updated, req.TaskID)
	if !ok {
		return Result{}, fmt.Errorf("task %s not found after validation", req.TaskID)
	}
	return Result{
		TaskDB:            updated,
		Task:              record,
		Validation:        commandResult,
		Evidence:          evidence,
		Receipt:           receipt,
		Transition:        transition,
		TransitionReceipt: transitionReceipt,
		Provider:          providerForRun,
		CommandID:         req.CommandID,
	}, nil
}
