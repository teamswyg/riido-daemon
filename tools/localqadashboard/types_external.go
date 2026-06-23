package main

type externalEvidenceFile struct {
	SchemaVersion string             `json:"schema_version"`
	ID            string             `json:"id"`
	ObservedAt    string             `json:"observed_at"`
	ExpiresAt     string             `json:"expires_at"`
	Status        string             `json:"status"`
	Scenarios     []externalScenario `json:"scenarios"`
	Evidence      string             `json:"-"`
}

type externalScenario struct {
	ID             string          `json:"id"`
	Status         string          `json:"status"`
	FailureSummary string          `json:"failure_summary,omitempty"`
	Screenshot     string          `json:"screenshot,omitempty"`
	Repair         *repairEvidence `json:"repair,omitempty"`
	Evidence       string          `json:"evidence,omitempty"`
	ExpiresAt      string          `json:"expires_at,omitempty"`
}
