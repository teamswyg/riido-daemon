package main

type Manifest struct {
	SchemaVersion    string            `json:"schema_version"`
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	GeneratedDoc     string            `json:"generated_doc"`
	Workflow         string            `json:"workflow"`
	EvidenceArtifact string            `json:"evidence_artifact"`
	Purpose          string            `json:"purpose"`
	Inputs           []GateInput       `json:"inputs"`
	GateOrder        []GateStep        `json:"gate_order"`
	Outputs          []string          `json:"outputs"`
	FailureSemantics []FailureSemantic `json:"failure_semantics"`
	SourceChecks     []SourceCheck     `json:"source_checks"`
	Assertions       []string          `json:"assertions"`
}

type GateInput struct {
	Name         string   `json:"name"`
	Owner        string   `json:"owner"`
	SourceChecks []string `json:"source_checks"`
}

type GateStep struct {
	Step         string   `json:"step"`
	Summary      string   `json:"summary"`
	SourceChecks []string `json:"source_checks"`
}

type FailureSemantic struct {
	Case         string   `json:"case"`
	Meaning      string   `json:"meaning"`
	SourceChecks []string `json:"source_checks"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
