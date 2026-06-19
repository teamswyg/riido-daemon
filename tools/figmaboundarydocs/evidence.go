package main

type evidence struct {
	SchemaVersion    string   `json:"schema_version"`
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	GeneratedDocs    []string `json:"generated_docs"`
	EntryCount       int      `json:"entry_count"`
	ProblemSummaries []string `json:"problem_summaries"`
	EvidenceArtifact string   `json:"evidence_artifact"`
}

func buildEvidence(m manifest, problems []problem) evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return evidence{
		SchemaVersion:    "riido-figma-boundary-docs-result.v1",
		ID:               m.ID,
		Status:           status,
		GeneratedDocs:    generatedDocPaths(),
		EntryCount:       len(m.Entries),
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: "figma-boundary-docs",
	}
}
