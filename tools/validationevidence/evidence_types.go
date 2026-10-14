package main

type Evidence struct {
	SchemaVersion string              `json:"schema_version"`
	ID            string              `json:"id"`
	Status        string              `json:"status"`
	SourceChecks  []SourceCheckResult `json:"source_checks"`
	AbsentChecks  []AbsentCheck       `json:"absent_checks"`
	Problems      []string            `json:"problems"`
}

type SourceCheckResult struct {
	Name string `json:"name"`
	File string `json:"file"`
	Pass bool   `json:"pass"`
}

type AbsentCheck struct {
	Name string   `json:"name"`
	Pass bool     `json:"pass"`
	Hits []string `json:"hits"`
}

type problem struct {
	Message string
}
