package main

type Manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	Purpose          string        `json:"purpose"`
	Surfaces         []Surface     `json:"surfaces"`
	SourceChecks     []SourceCheck `json:"source_checks"`
	Assertions       []string      `json:"assertions"`
}

type Surface struct {
	Provider     string   `json:"provider"`
	Surface      string   `json:"surface"`
	Flag         string   `json:"flag"`
	Enforcement  string   `json:"enforcement"`
	HostDefault  string   `json:"host_default"`
	SourceChecks []string `json:"source_checks"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
