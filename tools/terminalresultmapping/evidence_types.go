package main

type Evidence struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Status           string          `json:"status"`
	MappingChecks    []MappingCheck  `json:"mapping_checks"`
	CoverageChecks   []CoverageCheck `json:"coverage_checks"`
	DefaultChecks    []DefaultCheck  `json:"default_checks"`
	Assertions       []string        `json:"assertions"`
	ProblemSummaries []string        `json:"problem_summaries"`
	EvidenceArtifact string          `json:"evidence_artifact"`
}

type MappingCheck struct {
	StatusConst        string `json:"status_const"`
	Status             string `json:"status"`
	ExpectedEventConst string `json:"expected_event_const"`
	ActualEventConst   string `json:"actual_event_const"`
	ActualResolution   string `json:"actual_resolution"`
	ExpectedEventValue string `json:"expected_event_value"`
	ContractEventValue string `json:"contract_event_value"`
	Pass               bool   `json:"pass"`
}

type CoverageCheck struct {
	StatusConst string `json:"status_const"`
	Status      string `json:"status"`
	Covered     bool   `json:"covered"`
}

type DefaultCheck struct {
	Name     string `json:"name"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Pass     bool   `json:"pass"`
}
