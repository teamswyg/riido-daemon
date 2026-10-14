package project

import "github.com/teamswyg/riido-daemon/internal/taskdb"

func taskDBProviderCandidates(candidates []ProviderCandidate) []taskdb.ProviderCandidate {
	out := make([]taskdb.ProviderCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		out = append(out, taskdb.ProviderCandidate{
			ID:               candidate.ID,
			SourceWorkflow:   candidate.SourceWorkflow,
			Available:        candidate.Available,
			ApprovalRequired: candidate.ApprovalRequired,
		})
	}
	return out
}

func taskDBDiagnostics(diagnostics []ProjectionDiagnostic) []taskdb.ProjectionDiagnostic {
	out := make([]taskdb.ProjectionDiagnostic, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		out = append(out, taskdb.ProjectionDiagnostic{
			Severity: diagnostic.Severity,
			Code:     diagnostic.Code,
			Message:  diagnostic.Message,
		})
	}
	return out
}
