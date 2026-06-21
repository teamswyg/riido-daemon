package main

type manifest struct {
	SchemaVersion   string         `json:"schema_version"`
	ID              string         `json:"id"`
	LoopSource      string         `json:"loop_source"`
	WorkflowSources []string       `json:"workflow_sources,omitempty"`
	Workflows       []workflowSpec `json:"workflows"`
}

type workflowSpec struct {
	ID                string   `json:"id"`
	Workflow          string   `json:"workflow"`
	EvidenceArtifact  string   `json:"evidence_artifact"`
	RequiredFragments []string `json:"required_fragments"`
}

type evidence struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	Workflow         string           `json:"workflow"`
	EvidenceArtifact string           `json:"evidence_artifact"`
	MatchedCount     int              `json:"matched_count"`
	RequiredCount    int              `json:"required_count"`
	Fragments        []fragmentResult `json:"fragments"`
	LoopSource       string           `json:"loop_source"`
}

type fragmentResult struct {
	Value string `json:"value"`
	Found bool   `json:"found"`
}
