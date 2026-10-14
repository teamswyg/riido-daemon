package main

type evidence struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Status           string        `json:"status"`
	GeneratedDocs    []string      `json:"generated_docs"`
	PackageChecks    []checkResult `json:"package_checks"`
	ImportChecks     []checkResult `json:"import_checks"`
	ProblemSummaries []string      `json:"problem_summaries"`
	EvidenceArtifact string        `json:"evidence_artifact"`
}

func buildEvidence(m manifest, problems []problem, packageChecks, importChecks []checkResult) evidence {
	status := "verified"
	if len(problems) > 0 {
		status = "failed"
	}
	return evidence{
		SchemaVersion:    "riido-module-decomposition-result.v1",
		ID:               m.ID,
		Status:           status,
		GeneratedDocs:    generatedDocPaths(m),
		PackageChecks:    packageChecks,
		ImportChecks:     importChecks,
		ProblemSummaries: problemMessages(problems),
		EvidenceArtifact: m.EvidenceArtifact,
	}
}
