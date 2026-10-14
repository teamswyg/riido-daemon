package main

type Manifest struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	Workflow         string          `json:"workflow"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	Steps            []Step          `json:"steps"`
	SourceChecks     []SourceCheck   `json:"source_checks"`
	AbsentSurfaces   []AbsentSurface `json:"absent_surfaces"`
	Assertions       []string        `json:"assertions"`
}

type Step struct {
	Step           string   `json:"step"`
	Status         string   `json:"status"`
	Responsibility string   `json:"responsibility"`
	SourceChecks   []string `json:"source_checks"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type AbsentSurface struct {
	Name   string   `json:"name"`
	Scope  []string `json:"scope"`
	Tokens []string `json:"tokens"`
	Reason string   `json:"reason"`
}
