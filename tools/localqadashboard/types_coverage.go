package main

type coverageManifest struct {
	SchemaVersion string             `json:"schema_version"`
	ID            string             `json:"id"`
	Title         string             `json:"title"`
	Scenarios     []coverageScenario `json:"scenarios"`
}

type coverageScenario struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Tier       string `json:"tier"`
	Surface    string `json:"surface"`
	Evidence   string `json:"evidence"`
	ProviderID string `json:"provider_id,omitempty"`
}

type coverageRow struct {
	ID         string         `json:"id"`
	Title      string         `json:"title"`
	Tier       string         `json:"tier"`
	Surface    string         `json:"surface"`
	Status     string         `json:"status"`
	Evidence   string         `json:"evidence,omitempty"`
	ExpiresAt  string         `json:"expires_at,omitempty"`
	Repair     repairEvidence `json:"-"`
	Detail     string         `json:"detail,omitempty"`
	Screenshot string         `json:"screenshot,omitempty"`
}

type coverageSummary struct {
	Total       int `json:"total"`
	Passed      int `json:"passed"`
	Skipped     int `json:"skipped"`
	NotVerified int `json:"not_verified"`
	Failed      int `json:"failed"`
}
