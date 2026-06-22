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
	ID         string
	Title      string
	Tier       string
	Surface    string
	Status     string
	Repair     repairEvidence
	Detail     string
	Screenshot string
}

type coverageSummary struct {
	Total       int
	Passed      int
	Skipped     int
	NotVerified int
	Failed      int
}
