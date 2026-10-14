package main

import (
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/taskvalidation"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func printTaskValidationResult(taskDBPath string, result taskvalidation.Result) error {
	return printJSON(struct {
		OK                bool                             `json:"ok"`
		TaskDBPath        string                           `json:"task_db_path"`
		Validation        validation.CommandResult         `json:"validation"`
		Evidence          taskdb.TaskEvidenceRecord        `json:"evidence"`
		Receipt           taskdb.TaskCommandReceiptRecord  `json:"receipt"`
		Transition        *taskdb.TaskTransitionRecord     `json:"transition,omitempty"`
		TransitionReceipt *taskdb.TaskCommandReceiptRecord `json:"transition_receipt,omitempty"`
	}{
		OK:                result.Evidence.Result == "passed",
		TaskDBPath:        taskDBPath,
		Validation:        result.Validation,
		Evidence:          result.Evidence,
		Receipt:           result.Receipt,
		Transition:        result.Transition,
		TransitionReceipt: result.TransitionReceipt,
	})
}
