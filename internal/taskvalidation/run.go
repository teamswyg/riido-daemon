package taskvalidation

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
)

func Run(ctx context.Context, db taskdb.TaskDB, req Request, now time.Time) (Result, error) {
	req = req.normalized()
	if err := req.validate(); err != nil {
		return Result{}, err
	}
	if now.IsZero() {
		now = time.Now()
	}
	if ctx == nil {
		ctx = context.Background()
	}

	taskBeforeValidation, ok := FindTask(db, req.TaskID)
	if !ok {
		return Result{}, fmt.Errorf("task %s not found", req.TaskID)
	}
	providerForRun, err := ProviderForTask(db, req.TaskID, req.Provider)
	if err != nil {
		return Result{}, err
	}
	if err := ValidateDecisionLLMForTask(db, req.TaskID, req.DecisionLLM); err != nil {
		return Result{}, err
	}
	if req.CommandID == "" {
		req.CommandID = CommandID(req.TaskID, now)
	}

	commandResult, err := validation.RunCommand(ctx, validation.CommandRequest{
		Command:        req.Command,
		Workdir:        req.Workdir,
		Timeout:        req.Timeout,
		CommandID:      req.CommandID,
		Provider:       providerForRun,
		ValidationGate: req.ValidationGate,
		Summary:        req.Summary,
	}, now)
	if err != nil {
		return Result{}, err
	}
	updated, evidence, receipt, err := addValidationEvidence(db, req, providerForRun, commandResult, now)
	if err != nil {
		return Result{}, err
	}
	transition, transitionReceipt, err := maybeApplyValidationTransition(
		&updated, taskBeforeValidation, req, providerForRun, commandResult, now,
	)
	if err != nil {
		return Result{}, err
	}

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
