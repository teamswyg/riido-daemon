package main

type Evidence struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Status           string        `json:"status"`
	EnvConstChecks   []CheckResult `json:"env_const_checks"`
	AnchorChecks     []CheckResult `json:"anchor_checks"`
	SourceChecks     []CheckResult `json:"source_checks"`
	EnvVarCount      int           `json:"env_var_count"`
	ProblemSummaries []string      `json:"problem_summaries"`
	EvidenceArtifact string        `json:"evidence_artifact"`
}

type CheckResult struct {
	Name   string `json:"name"`
	File   string `json:"file,omitempty"`
	Pass   bool   `json:"pass"`
	Detail string `json:"detail,omitempty"`
}
