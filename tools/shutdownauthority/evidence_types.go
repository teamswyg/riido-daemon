package main

type Evidence struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Status           string          `json:"status"`
	LevelChecks      []LevelCheck    `json:"level_checks"`
	TimeoutChecks    []TimeoutCheck  `json:"timeout_checks"`
	ConsumerChecks   []ConsumerCheck `json:"consumer_checks"`
	Assertions       []string        `json:"assertions"`
	ProblemSummaries []string        `json:"problem_summaries"`
	EvidenceArtifact string          `json:"evidence_artifact"`
}

type LevelCheck struct {
	Const         string `json:"const"`
	ExpectedName  string `json:"expected_name"`
	ActualName    string `json:"actual_name"`
	ExpectedOrder int    `json:"expected_order"`
	ActualOrder   int    `json:"actual_order"`
	Pass          bool   `json:"pass"`
}

type TimeoutCheck struct {
	Const            string `json:"const"`
	ExpectedDuration string `json:"expected_duration"`
	ActualDuration   string `json:"actual_duration"`
	Pass             bool   `json:"pass"`
}

type ConsumerCheck struct {
	File     string `json:"file"`
	Contains string `json:"contains"`
	Reason   string `json:"reason"`
	Pass     bool   `json:"pass"`
}
