package main

type Evidence struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Status           string         `json:"status"`
	ImplementedRules int            `json:"implemented_rules"`
	ReservedRules    int            `json:"reserved_rules"`
	SourceChecks     []SourceResult `json:"source_checks"`
	Reserved         []ReservedRule `json:"reserved"`
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

type ReservedRule struct {
	Section          string   `json:"section"`
	Name             string   `json:"name"`
	RequiredEvidence string   `json:"required_evidence"`
	DecisionRefs     []string `json:"decision_refs,omitempty"`
}
