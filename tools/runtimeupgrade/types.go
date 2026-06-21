package main

type Manifest struct {
	SchemaVersion    string          `json:"schema_version"`
	LoopSource       string          `json:"loop_source,omitempty"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	Workflow         string          `json:"workflow"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	Invariant        string          `json:"invariant"`
	Inputs           []Rule          `json:"inputs"`
	Flow             []Rule          `json:"flow"`
	Policies         []Rule          `json:"policies"`
	NativeConfig     []Rule          `json:"native_config"`
	AbsentSurfaces   []AbsentSurface `json:"absent_surfaces,omitempty"`
	SourceChecks     []SourceCheck   `json:"source_checks"`
	Assertions       []string        `json:"assertions"`
}

type Rule struct {
	Name             string   `json:"name,omitempty"`
	Step             string   `json:"step,omitempty"`
	Owner            string   `json:"owner,omitempty"`
	Status           string   `json:"status"`
	DecisionRefs     []string `json:"decision_refs,omitempty"`
	Summary          string   `json:"summary,omitempty"`
	SourceChecks     []string `json:"source_checks,omitempty"`
	RequiredEvidence string   `json:"required_evidence,omitempty"`
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
