package main

type Evidence struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	AllowedChecks    []AllowedCheck   `json:"allowed_checks"`
	ForbiddenChecks  []ForbiddenCheck `json:"forbidden_checks"`
	Assertions       []string         `json:"assertions"`
	ProblemSummaries []string         `json:"problem_summaries"`
	EvidenceArtifact string           `json:"evidence_artifact"`
}

type AllowedCheck struct {
	Field    string `json:"field"`
	Status   string `json:"status"`
	Source   string `json:"source,omitempty"`
	Contains string `json:"contains,omitempty"`
	Pass     bool   `json:"pass"`
}

type ForbiddenCheck struct {
	Field  string   `json:"field"`
	Scope  []string `json:"scope"`
	Tokens []string `json:"tokens"`
	Hits   []string `json:"hits"`
	Pass   bool     `json:"pass"`
}
