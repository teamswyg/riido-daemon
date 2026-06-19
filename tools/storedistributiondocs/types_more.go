package main

type externalSource struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type workRow struct {
	Work         string `json:"work"`
	OwnerContext string `json:"owner_context"`
	Output       string `json:"output"`
}

type evidence struct {
	SchemaVersion    string   `json:"schema_version"`
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	GeneratedDocs    []string `json:"generated_docs"`
	Channels         []string `json:"channels"`
	ProblemSummaries []string `json:"problem_summaries,omitempty"`
	EvidenceArtifact string   `json:"evidence_artifact"`
}
