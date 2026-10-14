package taskvalidation

import (
	"context"
	"fmt"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
)

func Run(ctx context.Context, db taskdb.TaskDB, req Request, now time.Time) (Result, error) {
	req = req.normalized()
	if err := req.validate(); err != nil {
		return Result{}, err
	}
	ctx, now = normalizeRunContext(ctx, now)

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

	commandResult, err := runValidationCommand(ctx, req, providerForRun, now)
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

	return buildValidationResult(
		updated, req, commandResult, evidence, receipt, transition, transitionReceipt, providerForRun,
	)
}
