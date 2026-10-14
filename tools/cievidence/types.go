package main

type options struct {
	Workflow    string
	ID          string
	Manifest    string
	EvidenceOut string
}

type evidence struct {
	SchemaVersion    string     `json:"schema_version"`
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	Workflow         string     `json:"workflow"`
	EvidenceArtifact string     `json:"evidence_artifact"`
	Required         []required `json:"required_commands"`
	ProblemCount     int        `json:"problem_count"`
	Problems         []string   `json:"problem_summaries"`
	LoopSource       string     `json:"loop_source"`
}

type required struct {
	Command string `json:"command"`
	Found   bool   `json:"found"`
}

type manifest struct {
	SchemaVersion string         `json:"schema_version"`
	ID            string         `json:"id"`
	LoopSource    string         `json:"loop_source"`
	Workflows     []workflowSpec `json:"workflows"`
}

type workflowSpec struct {
	ID               string   `json:"id"`
	Workflow         string   `json:"workflow"`
	EvidenceArtifact string   `json:"evidence_artifact"`
	RequiredCommands []string `json:"required_commands"`
}
