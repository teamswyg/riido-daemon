package main

type evidence struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Status           string        `json:"status"`
	GeneratedDocs    []string      `json:"generated_docs"`
	Targets          []string      `json:"targets"`
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
		SchemaVersion:    "riido-release-artifacts-result.v1",
		ID:               m.ID,
		Status:           status,
		GeneratedDocs:    generatedDocPaths(m),
		Targets:          targetKeys(m.Targets),
		SourceChecks:     sourceChecks,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: m.EvidenceArtifact,
	}
}

func targetKeys(targets []target) []string {
	keys := make([]string, 0, len(targets))
	for _, target := range targets {
		keys = append(keys, target.GOOS+"/"+target.GOARCH+"/"+target.Format)
	}
	return keys
}
