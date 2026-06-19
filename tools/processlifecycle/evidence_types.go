package main

type Evidence struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	InterfaceChecks  []InterfaceCheck `json:"interface_checks"`
	SourceChecks     []SourceResult   `json:"source_checks"`
	Assertions       []string         `json:"assertions"`
	ProblemSummaries []string         `json:"problem_summaries"`
	EvidenceArtifact string           `json:"evidence_artifact"`
}

type InterfaceCheck struct {
	Interface string   `json:"interface"`
	File      string   `json:"file"`
	Required  []string `json:"required"`
	Actual    []string `json:"actual"`
	Pass      bool     `json:"pass"`
}

type SourceResult struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
	Pass     bool   `json:"pass"`
}
