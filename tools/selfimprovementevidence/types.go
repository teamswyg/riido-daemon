package main

type options struct {
	Manifest    string
	EvidenceDir string
	EvidenceOut string
	WriteDoc    bool
	CheckDoc    bool
}

type manifest struct {
	SchemaVersion    string             `json:"schema_version"`
	ID               string             `json:"id"`
	Title            string             `json:"title"`
	GeneratedDoc     string             `json:"generated_doc"`
	Workflow         string             `json:"workflow"`
	EvidenceArtifact string             `json:"evidence_artifact"`
	LoopSource       string             `json:"loop_source"`
	Required         []requiredEvidence `json:"required_evidence"`
	ClosedLoops      []closedLoop       `json:"closed_loop_classes"`
}

type requiredEvidence struct {
	ID          string      `json:"id"`
	File        string      `json:"file"`
	Description string      `json:"description"`
	Producer    string      `json:"producer_command"`
	Assertions  []assertion `json:"assertions"`
}

type assertion struct {
	Field  string `json:"field"`
	Equals any    `json:"equals,omitempty"`
	Empty  bool   `json:"empty,omitempty"`
}

type closedLoop struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	EvidenceIDs []string `json:"evidence_ids"`
}

type report struct {
	SchemaVersion  string          `json:"schema_version"`
	ID             string          `json:"id"`
	Status         string          `json:"status"`
	GeneratedDoc   string          `json:"generated_doc"`
	Workflow       string          `json:"workflow"`
	Artifact       string          `json:"evidence_artifact"`
	LoopSource     string          `json:"loop_source"`
	RequiredCount  int             `json:"required_evidence_count"`
	VerifiedCount  int             `json:"verified_evidence_count"`
	ClosedCount    int             `json:"closed_loop_count"`
	ClosedVerified int             `json:"verified_closed_loop_count"`
	CheckCount     int             `json:"check_count"`
	PassingCount   int             `json:"passing_check_count"`
	ProblemCount   int             `json:"problem_count"`
	Problems       []string        `json:"problem_summaries"`
	Checks         []checkSummary  `json:"checks"`
	ClosedLoops    []closedSummary `json:"closed_loop_classes"`
}

type checkSummary struct {
	EvidenceID string `json:"evidence_id"`
	Field      string `json:"field"`
	Status     string `json:"status"`
}

type closedSummary struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
}
