package main

type Manifest struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	Workflow         string          `json:"workflow"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	Steps            []LifecycleStep `json:"steps"`
	Interfaces       []InterfaceSpec `json:"interfaces"`
	SourceChecks     []SourceCheck   `json:"source_checks"`
	Assertions       []string        `json:"assertions"`
}

type LifecycleStep struct {
	Step           string `json:"step"`
	Responsibility string `json:"responsibility"`
}

type InterfaceSpec struct {
	Name    string   `json:"name"`
	File    string   `json:"file"`
	Methods []string `json:"methods"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
