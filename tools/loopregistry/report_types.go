package main

type report struct {
	SchemaVersion      string          `json:"schema_version"`
	ID                 string          `json:"id"`
	Status             string          `json:"status"`
	GeneratedDoc       string          `json:"generated_doc"`
	Workflow           string          `json:"workflow"`
	EvidenceArtifact   string          `json:"evidence_artifact"`
	LoopCount          int             `json:"loop_count"`
	BusinessClaimCount int             `json:"business_claim_count"`
	ProblemCount       int             `json:"problem_count"`
	Problems           []string        `json:"problem_summaries"`
	Loops              []loopSummary   `json:"loops"`
	BusinessClaims     []claimSummary  `json:"business_claims"`
	ChangedFileCheck   *changedSummary `json:"changed_file_check,omitempty"`
}

type loopSummary struct {
	ID           string   `json:"id"`
	Kind         string   `json:"kind"`
	ExpiresAfter string   `json:"expires_after"`
	Evidence     []string `json:"evidence"`
}

type claimSummary struct {
	ID            string   `json:"id"`
	FileCount     int      `json:"file_count"`
	DocCount      int      `json:"doc_count"`
	EvidenceCount int      `json:"evidence_count"`
	VerifierCount int      `json:"verifier_count"`
	BoundFiles    []string `json:"bound_files"`
}

type changedSummary struct {
	InputCount        int      `json:"input_count"`
	MatchedClaimCount int      `json:"matched_claim_count"`
	MatchedClaims     []string `json:"matched_claims"`
	Problems          []string `json:"problems"`
}
