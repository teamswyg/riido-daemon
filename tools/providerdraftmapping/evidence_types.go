package main

type Evidence struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Status           string          `json:"status"`
	MappingChecks    []MappingCheck  `json:"mapping_checks"`
	CoverageChecks   []CoverageCheck `json:"coverage_checks"`
	Assertions       []string        `json:"assertions"`
	ProblemSummaries []string        `json:"problem_summaries,omitempty"`
	EvidenceArtifact string          `json:"evidence_artifact"`
}

type MappingCheck struct {
	EventKind string `json:"event_kind"`
	Expected  string `json:"expected"`
	Actual    string `json:"actual"`
	Passed    bool   `json:"passed"`
}

type CoverageCheck struct {
	EventKind string `json:"event_kind"`
	Category  string `json:"category"`
	Covered   bool   `json:"covered"`
}
