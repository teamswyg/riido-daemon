package main

type Evidence struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Status           string        `json:"status"`
	SourceChecks     []CheckResult `json:"source_checks"`
	ForbiddenChecks  []CheckResult `json:"forbidden_checks"`
	BehaviorChecks   []CheckResult `json:"behavior_checks"`
	CommandGroups    []string      `json:"command_groups"`
	Providers        []string      `json:"providers"`
	Assertions       []string      `json:"assertions"`
	ProblemSummaries []string      `json:"problem_summaries"`
	EvidenceArtifact string        `json:"evidence_artifact"`
}

type CheckResult struct {
	Name    string `json:"name"`
	File    string `json:"file,omitempty"`
	Command string `json:"command,omitempty"`
	Pass    bool   `json:"pass"`
	Detail  string `json:"detail,omitempty"`
}
