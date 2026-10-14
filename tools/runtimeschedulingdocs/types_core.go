package main

type coreDoc struct {
	SchemaVersion       string      `json:"schema_version"`
	ID                  string      `json:"id"`
	Title               string      `json:"title"`
	LoopSource          string      `json:"loop_source,omitempty"`
	GeneratedDoc        string      `json:"generated_doc"`
	Context             string      `json:"context"`
	Responsibilities    []string    `json:"responsibilities"`
	NonResponsibilities []string    `json:"non_responsibilities"`
	Invariants          []invariant `json:"invariants"`
}

type invariant struct {
	Name         string   `json:"name"`
	Summary      string   `json:"summary"`
	SourceChecks []string `json:"source_checks,omitempty"`
}
