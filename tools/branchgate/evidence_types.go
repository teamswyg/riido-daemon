package main

type Evidence struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Status           string         `json:"status"`
	ScriptChecks     []ScriptCheck  `json:"script_checks"`
	ExampleChecks    []ExampleCheck `json:"example_checks"`
	Assertions       []string       `json:"assertions"`
	ProblemSummaries []string       `json:"problem_summaries"`
	EvidenceArtifact string         `json:"evidence_artifact"`
}

type ScriptCheck struct {
	Name string `json:"name"`
	File string `json:"file"`
	Pass bool   `json:"pass"`
}

type ExampleCheck struct {
	Branch       string `json:"branch"`
	WantAccepted bool   `json:"want_accepted"`
	ExitCode     int    `json:"exit_code"`
	Pass         bool   `json:"pass"`
}
