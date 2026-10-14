package main

type manifest struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	GeneratedDoc     string           `json:"generated_doc"`
	Workflow         string           `json:"workflow"`
	EvidenceArtifact string           `json:"evidence_artifact"`
	EvidenceTool     string           `json:"evidence_tool"`
	Decisions        []decisionRecord `json:"decisions"`
	Assertions       []string         `json:"assertions"`
	Loop             evidenceLoop     `json:"loop"`
}

type decisionRecord struct {
	CandidateID    string `json:"candidate_id"`
	CandidateScope string `json:"candidate_scope,omitempty"`
	Disposition    string `json:"disposition"`
	Priority       string `json:"priority"`
	Owner          string `json:"owner"`
	NextLoop       string `json:"next_loop"`
	NextArtifact   string `json:"next_artifact"`
	ReviewBy       string `json:"review_by,omitempty"`
	Reason         string `json:"reason"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
