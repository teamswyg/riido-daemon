package main

type Evidence struct {
	ID             string                `json:"id"`
	SchemaVersion  string                `json:"schema_version"`
	GeneratedDoc   string                `json:"generated_doc"`
	Workflow       string                `json:"workflow"`
	SourcePackage  string                `json:"source_package"`
	Problems       []problem             `json:"problems"`
	FSM            FSMSnapshot           `json:"fsm"`
	SourceChecks   []SourceCheckEvidence `json:"source_checks"`
	ForbiddenCheck ForbiddenCheck        `json:"forbidden_doc_tokens"`
}

type SourceCheckEvidence struct {
	Name string `json:"name"`
	File string `json:"file"`
	OK   bool   `json:"ok"`
}

type ForbiddenCheck struct {
	Tokens []string `json:"tokens"`
	OK     bool     `json:"ok"`
}
