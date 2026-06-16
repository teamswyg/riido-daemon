package riidoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/teamswyg/riido-contracts/task"
	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/validation"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func (s Server) validateTask(ctx context.Context, params json.RawMessage) (ValidateResponse, error) {
	var req ValidateRequest
	if len(params) == 0 {
		return ValidateResponse{}, errors.New("validate params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return ValidateResponse{}, fmt.Errorf("decode validate params: %w", err)
	}
	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		return ValidateResponse{}, errors.New("task_id is required")
	}
	if strings.TrimSpace(req.Command) == "" {
		return ValidateResponse{}, errors.New("command is required")
	}
	if strings.TrimSpace(req.ApprovalID) == "" {
		return ValidateResponse{}, errors.New("approval_id is required before validation command execution")
	}
	if req.TimeoutSeconds < 0 {
		return ValidateResponse{}, errors.New("timeout_seconds must not be negative")
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return ValidateResponse{}, err
	}
	providerForRun, err := validationProviderForTask(db, taskID, req.Provider)
	if err != nil {
		return ValidateResponse{}, err
	}
	if err := validateDecisionLLMForTask(db, taskID, req.DecisionLLM); err != nil {
		return ValidateResponse{}, err
	}
	taskBeforeValidation, ok := findTask(db, taskID)
	if !ok {
		return ValidateResponse{}, fmt.Errorf("task %s not found", taskID)
	}

	now := time.Now()
	commandID := strings.TrimSpace(req.CommandID)
	if commandID == "" {
		commandID = validationCommandID(taskID, now)
	}
	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	actor := textutil.Default(req.Actor, "daemon")
	source := textutil.Default(req.Source, "riido-api")
	result, err := validation.RunCommand(ctx, validation.CommandRequest{
		Command:        req.Command,
		Workdir:        req.Workdir,
		Timeout:        timeout,
		CommandID:      commandID,
		Provider:       providerForRun,
		ValidationGate: req.ValidationGate,
		Summary:        req.Summary,
	}, now)
	if err != nil {
		return ValidateResponse{}, err
	}
	updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
		TaskID:            taskID,
		Command:           result.Command,
		ExitCode:          result.ExitCode,
		Result:            result.Result,
		Actor:             actor,
		Source:            source,
		Summary:           result.Summary,
		ValidationGate:    result.ValidationGate,
		ProviderRunID:     result.ProviderRunID,
		ProviderRunResult: result.ProviderRunResult,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   commandID,
			Provider:    providerForRun,
			DecisionLLM: req.DecisionLLM,
			ApprovalID:  req.ApprovalID,
		},
	}, now)
	if err != nil {
		return ValidateResponse{}, err
	}

	var transition *taskdb.TaskTransitionRecord
	var transitionReceipt *taskdb.TaskCommandReceiptRecord
	if taskBeforeValidation.State == task.StateValidating {
		toState, eventType := validationTransitionForResult(result.Result)
		nextDB, appliedTransition, appliedReceipt, err := taskdb.ApplyGuardedTaskTransition(updated, taskdb.TaskTransitionInput{
			TaskID:  taskID,
			ToState: toState,
			Event:   eventType,
			Actor:   actor,
			Source:  source,
			Reason:  fmt.Sprintf("validation %s via %s", result.Result, result.ValidationGate),
			Guard: taskdb.TaskMutationGuardInput{
				CommandID:   commandID + ":transition",
				Provider:    providerForRun,
				DecisionLLM: req.DecisionLLM,
				ApprovalID:  req.ApprovalID,
			},
		}, now)
		if err != nil {
			return ValidateResponse{}, err
		}
		updated = nextDB
		transition = &appliedTransition
		transitionReceipt = &appliedReceipt
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, updated); err != nil {
		return ValidateResponse{}, err
	}
	record, ok := findTask(updated, taskID)
	if !ok {
		return ValidateResponse{}, fmt.Errorf("task %s not found after validation", taskID)
	}
	return ValidateResponse{
		TaskDBPath:        s.config.TaskDBPath,
		Task:              record,
		Validation:        result,
		Evidence:          evidence,
		Receipt:           receipt,
		Transition:        transition,
		TransitionReceipt: transitionReceipt,
	}, nil
}

func (s Server) addEvidence(params json.RawMessage) (EvidenceResponse, error) {
	var req EvidenceRequest
	if len(params) == 0 {
		return EvidenceResponse{}, errors.New("evidence params are required")
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return EvidenceResponse{}, fmt.Errorf("decode evidence params: %w", err)
	}
	db, err := taskdb.LoadTaskDB(s.config.TaskDBPath)
	if err != nil {
		return EvidenceResponse{}, err
	}
	updated, evidence, receipt, err := taskdb.AddGuardedTaskEvidence(db, taskdb.TaskEvidenceInput{
		TaskID:            req.TaskID,
		Command:           req.Command,
		ExitCode:          req.ExitCode,
		Result:            req.Result,
		Actor:             req.Actor,
		Source:            req.Source,
		Summary:           req.Summary,
		ValidationGate:    req.ValidationGate,
		ProviderRunID:     req.ProviderRunID,
		ProviderRunResult: req.ProviderRunResult,
		Guard: taskdb.TaskMutationGuardInput{
			CommandID:   req.CommandID,
			Provider:    req.Provider,
			DecisionLLM: req.DecisionLLM,
			ApprovalID:  req.ApprovalID,
		},
	}, time.Now())
	if err != nil {
		return EvidenceResponse{}, err
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, updated); err != nil {
		return EvidenceResponse{}, err
	}
	record, ok := findTask(updated, req.TaskID)
	if !ok {
		return EvidenceResponse{}, fmt.Errorf("task %s not found after evidence append", req.TaskID)
	}
	return EvidenceResponse{
		TaskDBPath: s.config.TaskDBPath,
		Task:       record,
		Evidence:   evidence,
		Receipt:    receipt,
	}, nil
}
