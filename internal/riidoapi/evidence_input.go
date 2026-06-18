package riidoapi

import "github.com/teamswyg/riido-daemon/internal/taskdb"

func evidenceInput(req EvidenceRequest) taskdb.TaskEvidenceInput {
	return taskdb.TaskEvidenceInput{
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
	}
}
