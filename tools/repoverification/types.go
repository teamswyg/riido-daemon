package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	Commands         []commandSpec `json:"commands"`
	Assertions       []string      `json:"assertions"`
}

type commandSpec struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Argv        []string `json:"argv"`
}

type evidenceFile struct {
	SchemaVersion string            `json:"schema_version"`
	ID            string            `json:"id"`
	ObservedAt    string            `json:"observed_at"`
	Status        string            `json:"status"`
	Commands      []commandEvidence `json:"commands"`
	Assertions    []string          `json:"assertions"`
}

type commandEvidence struct {
	ID         string `json:"id"`
	Argv       string `json:"argv"`
	Status     string `json:"status"`
	DurationMS int64  `json:"duration_ms"`
	OutputTail string `json:"output_tail,omitempty"`
}
