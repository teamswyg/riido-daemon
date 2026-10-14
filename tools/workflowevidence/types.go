package main

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	GeneratedDoc     string        `json:"generated_doc"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	WorkflowRoot     string        `json:"workflow_root"`
	AcceptedGaps     []acceptedGap `json:"accepted_gaps"`
	Assertions       []string      `json:"assertions"`
	Loop             evidenceLoop  `json:"loop"`
}

type acceptedGap struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	Next   string `json:"next"`
}

type workflowRecord struct {
	Path                 string   `json:"path"`
	Status               string   `json:"status"`
	HasExecutable        bool     `json:"has_executable"`
	HasEvidenceOut       bool     `json:"has_evidence_out"`
	EvidenceOutCount     int      `json:"evidence_out_count"`
	UploadedEvidenceOut  int      `json:"uploaded_evidence_out_count"`
	MissingEvidenceOut   []string `json:"missing_evidence_out_uploads,omitempty"`
	UploadsArtifact      bool     `json:"uploads_artifact"`
	ArtifactUploadCount  int      `json:"artifact_upload_count"`
	StrictUploadCount    int      `json:"strict_upload_count"`
	NonStrictUploadCount int      `json:"non_strict_upload_count"`
	Reason               string   `json:"reason,omitempty"`
	Next                 string   `json:"next,omitempty"`
}

type evidenceLoop struct {
	Observation   string `json:"observation"`
	Hypothesis    string `json:"hypothesis"`
	Execute       string `json:"execute"`
	Evaluate      string `json:"evaluate"`
	Retrospective string `json:"retrospective"`
}
