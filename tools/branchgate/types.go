package main

type Manifest struct {
	SchemaVersion    string          `json:"schema_version"`
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	GeneratedDoc     string          `json:"generated_doc"`
	GeneratedScript  string          `json:"generated_script"`
	Workflow         string          `json:"workflow"`
	EvidenceWorkflow string          `json:"evidence_workflow"`
	EvidenceArtifact string          `json:"evidence_artifact"`
	Pattern          string          `json:"pattern"`
	Shape            string          `json:"shape"`
	Example          string          `json:"example"`
	AllowMain        bool            `json:"allow_main"`
	Rules            []string        `json:"rules"`
	Examples         []BranchExample `json:"examples"`
	Assertions       []string        `json:"assertions"`
}

type BranchExample struct {
	Branch   string `json:"branch"`
	Accepted bool   `json:"accepted"`
	Reason   string `json:"reason"`
}
