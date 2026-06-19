package main

type Evidence struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Status           string         `json:"status"`
	GeneratedDocs    []string       `json:"generated_docs"`
	SourceChecks     []SourceResult `json:"source_checks"`
	AbsentChecks     []AbsentCheck  `json:"absent_checks"`
	Assertions       []string       `json:"assertions"`
	ProblemSummaries []string       `json:"problem_summaries"`
	EvidenceArtifact string         `json:"evidence_artifact"`
}

type SourceResult struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
	Pass     bool   `json:"pass"`
}

type AbsentCheck struct {
	Name   string   `json:"name"`
	Scope  []string `json:"scope"`
	Tokens []string `json:"tokens"`
	Hits   []string `json:"hits"`
	Pass   bool     `json:"pass"`
}
