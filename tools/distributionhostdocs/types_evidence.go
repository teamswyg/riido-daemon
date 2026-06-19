package main

type evidence struct {
	SchemaVersion    string              `json:"schema_version"`
	ID               string              `json:"id"`
	Status           string              `json:"status"`
	GeneratedDocs    []string            `json:"generated_docs"`
	SourceChecks     []sourceCheckResult `json:"source_checks"`
	AssertionCount   int                 `json:"assertion_count"`
	PageCount        int                 `json:"page_count"`
	ProblemSummaries []string            `json:"problem_summaries,omitempty"`
	EvidenceArtifact string              `json:"evidence_artifact"`
}

type sourceCheckResult struct {
	Name   string `json:"name"`
	File   string `json:"file"`
	Passed bool   `json:"passed"`
}
