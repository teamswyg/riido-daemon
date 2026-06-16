package riidoapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskdb"
	"github.com/teamswyg/riido-daemon/internal/taskvalidation"
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

	now := time.Now()
	result, err := taskvalidation.Run(ctx, db, taskvalidation.Request{
		TaskID:         taskID,
		Command:        req.Command,
		Workdir:        req.Workdir,
		Timeout:        time.Duration(req.TimeoutSeconds) * time.Second,
		Actor:          textutil.Default(req.Actor, "daemon"),
		Source:         textutil.Default(req.Source, "riido-api"),
		Summary:        req.Summary,
		Provider:       req.Provider,
		DecisionLLM:    req.DecisionLLM,
		ApprovalID:     req.ApprovalID,
		CommandID:      req.CommandID,
		ValidationGate: req.ValidationGate,
	}, now)
	if err != nil {
		return ValidateResponse{}, err
	}
	if err := taskdb.SaveTaskDB(s.config.TaskDBPath, result.TaskDB); err != nil {
		return ValidateResponse{}, err
	}
	return ValidateResponse{
		TaskDBPath:        s.config.TaskDBPath,
		Task:              result.Task,
		Validation:        result.Validation,
		Evidence:          result.Evidence,
		Receipt:           result.Receipt,
		Transition:        result.Transition,
		TransitionReceipt: result.TransitionReceipt,
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
