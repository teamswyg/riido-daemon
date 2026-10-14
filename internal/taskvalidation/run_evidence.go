package taskvalidation

import (
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func addValidationEvidence(
	db taskdb.TaskDB,
	req Request,
	providerForRun string,
	result validation.CommandResult,
	now time.Time,
) (taskdb.TaskDB, taskdb.TaskEvidenceRecord, taskdb.TaskCommandReceiptRecord, error) {
	return taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
		TaskID:            req.TaskID,
		Command:           result.Command,
		ExitCode:          result.ExitCode,
		Result:            result.Result,
		Actor:             req.Actor,
		Source:            req.Source,
		Summary:           result.Summary,
		ValidationGate:    result.ValidationGate,
		ProviderRunID:     result.ProviderRunID,
		ProviderRunResult: result.ProviderRunResult,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   req.CommandID,
			Provider:    providerForRun,
			DecisionLLM: req.DecisionLLM,
			ApprovalID:  req.ApprovalID,
		},
	}, now)
}
