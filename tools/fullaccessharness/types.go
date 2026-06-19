package main

type Manifest struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	Workflow         string          `json:"workflow"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	Purpose          string          `json:"purpose"`
	Facts            []Fact          `json:"facts"`
	AbsentSurfaces   []AbsentSurface `json:"absent_surfaces"`
	SourceChecks     []SourceCheck   `json:"source_checks"`
	Assertions       []string        `json:"assertions"`
}

type Fact struct {
	Name         string   `json:"name"`
	Summary      string   `json:"summary"`
	SourceChecks []string `json:"source_checks"`
}

type AbsentSurface struct {
	Name   string   `json:"name"`
	Scope  []string `json:"scope"`
	Tokens []string `json:"tokens"`
	Reason string   `json:"reason"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
