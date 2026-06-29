package main

type Manifest struct {
	SchemaVersion string    `json:"schema_version"`
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Bindings      []Binding `json:"bindings"`
}

type Binding struct {
	ID                   string   `json:"id"`
	Claim                string   `json:"claim"`
	ClaimClass           string   `json:"claim_class,omitempty"`
	Triggers             []string `json:"triggers"`
	RequiredWithTriggers []string `json:"required_with_triggers"`
	GeneratedDocs        []string `json:"generated_docs"`
	Verifiers            []string `json:"verifiers"`
	EvidenceIDs          []string `json:"evidence_ids"`
}

type options struct {
	Repo         string
	Manifest     string
	ChangedFiles []string
	EvidenceOut  string
}

type bindingResult struct {
	ID                   string   `json:"id"`
	Claim                string   `json:"claim"`
	ClaimClass           string   `json:"claim_class,omitempty"`
	Triggered            bool     `json:"triggered"`
	ChangedTriggers      []string `json:"changed_triggers"`
	MissingRequired      []string `json:"missing_required"`
	ChangedRequiredCount int      `json:"changed_required_count"`
}

type businessClaimSummary struct {
	Count         int      `json:"count"`
	VerifiedCount int      `json:"verified_count"`
	IDs           []string `json:"ids"`
}

type Evidence struct {
	SchemaVersion  string               `json:"schema_version"`
	Status         string               `json:"status"`
	ManifestID     string               `json:"manifest_id"`
	ChangedFiles   []string             `json:"changed_files"`
	BusinessClaims businessClaimSummary `json:"business_claims"`
	Results        []bindingResult      `json:"results"`
	ProblemCount   int                  `json:"problem_count"`
	Problems       []string             `json:"problems,omitempty"`
}
