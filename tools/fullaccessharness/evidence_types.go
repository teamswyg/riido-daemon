package main

type Evidence struct {
	ID             string                `json:"id"`
	SchemaVersion  string                `json:"schema_version"`
	GeneratedDoc   string                `json:"generated_doc"`
	Workflow       string                `json:"workflow"`
	Problems       []problem             `json:"problems"`
	SourceChecks   []SourceCheckEvidence `json:"source_checks"`
	AbsentSurfaces []AbsentEvidence      `json:"absent_surfaces"`
	Assertions     []string              `json:"assertions"`
}

type SourceCheckEvidence struct {
	Name string `json:"name"`
	File string `json:"file"`
	OK   bool   `json:"ok"`
}

type AbsentEvidence struct {
	Name  string   `json:"name"`
	Scope []string `json:"scope"`
	OK    bool     `json:"ok"`
}
