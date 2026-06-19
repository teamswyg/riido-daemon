package main

type Manifest struct {
	SchemaVersion       string   `json:"schema_version"`
	ID                  string   `json:"id"`
	Title               string   `json:"title"`
	GeneratedDoc        string   `json:"generated_doc"`
	Workflow            string   `json:"workflow"`
	EvidenceArtifact    string   `json:"evidence_artifact"`
	SemanticActivity    []string `json:"semantic_activity"`
	NonSemanticActivity []string `json:"non_semantic_activity"`
	Assertions          []string `json:"assertions"`
}

type Evidence struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	Classifications  []Classification `json:"classifications"`
	Assertions       []string         `json:"assertions"`
	ProblemSummaries []string         `json:"problem_summaries,omitempty"`
	EvidenceArtifact string           `json:"evidence_artifact"`
}

type Classification struct {
	Kind      string `json:"kind"`
	Manifest  string `json:"manifest"`
	Runtime   string `json:"runtime"`
	Confirmed bool   `json:"confirmed"`
}
