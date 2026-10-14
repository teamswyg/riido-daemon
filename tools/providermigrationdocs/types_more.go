package main

type link struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type evidence struct {
	SchemaVersion    string   `json:"schema_version"`
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	GeneratedDocs    []string `json:"generated_docs"`
	PageCount        int      `json:"page_count"`
	ProviderCount    int      `json:"provider_count"`
	ArtifactCount    int      `json:"artifact_count"`
	Assertions       []string `json:"assertions"`
	ProblemSummaries []string `json:"problem_summaries,omitempty"`
	EvidenceArtifact string   `json:"evidence_artifact"`
}
