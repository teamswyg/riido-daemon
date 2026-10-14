package main

type evidence struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Status           string        `json:"status"`
	GeneratedDocs    []string      `json:"generated_docs"`
	ProviderCount    int           `json:"provider_count"`
	SourceChecks     []checkResult `json:"source_checks"`
	ProblemSummaries []string      `json:"problem_summaries"`
	EvidenceArtifact string        `json:"evidence_artifact"`
}

func buildEvidence(m manifest, problems []problem, sourceChecks []checkResult) evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return evidence{
		SchemaVersion:    "riido-integration-matrix-docs-result.v1",
		ID:               m.ID,
		Status:           status,
		GeneratedDocs:    generatedDocPaths(m),
		ProviderCount:    len(m.ProviderValidation.Providers),
		SourceChecks:     sourceChecks,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: m.EvidenceArtifact,
	}
}
