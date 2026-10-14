package main

type model struct {
	Manifest   manifest
	Overview   overviewFragment
	Risk       riskFragment
	Execution  executionFragment
	Lifecycle  lifecycleFragment
	Governance governanceFragment
	Evidence   evidenceManifest
	Items      []evidenceItem
	Boundaries []boundaryItem
}

type result struct {
	SchemaVersion    string   `json:"schema_version"`
	ID               string   `json:"id"`
	Status           string   `json:"status"`
	GeneratedDocs    []string `json:"generated_docs"`
	EvidenceItems    int      `json:"evidence_items"`
	Remaining        int      `json:"remaining_boundaries"`
	ProblemSummaries []string `json:"problem_summaries,omitempty"`
	EvidenceArtifact string   `json:"evidence_artifact"`
}
