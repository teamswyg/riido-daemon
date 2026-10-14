package riidoapi

import (
	"strings"
	"time"

	"github.com/teamswyg/riido-daemon/internal/taskvalidation"
	"github.com/teamswyg/riido-daemon/pkg/util/textutil"
)

func validationRunRequest(req ValidateRequest) taskvalidation.Request {
	return taskvalidation.Request{
		TaskID:         strings.TrimSpace(req.TaskID),
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
	}
}

func validateResponse(path string, result taskvalidation.Result) ValidateResponse {
	return ValidateResponse{
		TaskDBPath:        path,
		Task:              result.Task,
		Validation:        result.Validation,
		Evidence:          result.Evidence,
		Receipt:           result.Receipt,
		Transition:        result.Transition,
		TransitionReceipt: result.TransitionReceipt,
	}
}
