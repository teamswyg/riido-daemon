package project

func appendDecisionGateDiagnostic(projection *WorkspaceProjection) {
	if projection.DecisionGate == "human-approval-required" {
		return
	}
	projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
		Severity: "error",
		Code:     "orchestration-human-gate-missing",
		Message:  "orchestration decision gate is " + projection.DecisionGate,
	})
}

func appendRecommendationDiagnostics(projection *WorkspaceProjection) {
	if projection.RecommendedProvider == "" {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "orchestration-recommended-provider-missing",
			Message:  "orchestration has no recommended provider",
		})
	}
	if projection.RecommendedDecisionLLM == "" {
		projection.Diagnostics = append(projection.Diagnostics, ProjectionDiagnostic{
			Severity: "error",
			Code:     "orchestration-recommended-decision-llm-missing",
			Message:  "orchestration has no recommended decision LLM",
		})
	}
}
