package main

type Manifest struct {
	SchemaVersion          string   `json:"schema_version"`
	ID                     string   `json:"id"`
	Title                  string   `json:"title"`
	GeneratedDoc           string   `json:"generated_doc"`
	Workflow               string   `json:"workflow"`
	EvidenceArtifact       string   `json:"evidence_artifact"`
	DraftSource            string   `json:"draft_source"`
	BuilderSource          string   `json:"builder_source"`
	DraftSuppliedFields    []string `json:"draft_supplied_fields"`
	IngestorAssignedFields []string `json:"ingestor_assigned_fields"`
	Rules                  []string `json:"rules"`
}

type Evidence struct {
	SchemaVersion    string       `json:"schema_version"`
	ID               string       `json:"id"`
	Status           string       `json:"status"`
	FieldChecks      []FieldCheck `json:"field_checks"`
	BuilderChecks    []FieldCheck `json:"builder_checks"`
	ProblemSummaries []string     `json:"problem_summaries,omitempty"`
	EvidenceArtifact string       `json:"evidence_artifact"`
}

type FieldCheck struct {
	Field    string `json:"field"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
}
