package main

type Evidence struct {
	ID             string                `json:"id"`
	SchemaVersion  string                `json:"schema_version"`
	GeneratedDoc   string                `json:"generated_doc"`
	Workflow       string                `json:"workflow"`
	PolicyArtifact string                `json:"policy_artifact"`
	Problems       []problem             `json:"problems"`
	Policy         PolicySnapshot        `json:"policy"`
	SourceChecks   []SourceCheckEvidence `json:"source_checks"`
	ShapeChecks    []ShapeCheck          `json:"shape_checks"`
}

type SourceCheckEvidence struct {
	Name string `json:"name"`
	File string `json:"file"`
	OK   bool   `json:"ok"`
}

type ShapeCheck struct {
	Name string `json:"name"`
	OK   bool   `json:"ok"`
}
