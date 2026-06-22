package main

type providerEvidenceFile struct {
	SchemaVersion   string             `json:"schema_version"`
	ID              string             `json:"id"`
	ObservedAt      string             `json:"observed_at"`
	ExpiresAt       string             `json:"expires_at"`
	FreshForSeconds int64              `json:"fresh_for_seconds"`
	Status          string             `json:"status"`
	RunIntegration  bool               `json:"run_integration"`
	Platform        evidencePlatform   `json:"platform"`
	Providers       []providerEvidence `json:"providers"`
}

type evidencePlatform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

type providerEvidence struct {
	ID                 string          `json:"id"`
	Available          bool            `json:"available"`
	ExecutableRef      string          `json:"executable_ref"`
	Version            string          `json:"version,omitempty"`
	IntegrationStatus  string          `json:"integration_status"`
	IntegrationCommand string          `json:"integration_command"`
	FailureSummary     string          `json:"failure_summary,omitempty"`
	Repair             *repairEvidence `json:"repair,omitempty"`
}

type repairEvidence struct {
	ProviderID       string `json:"provider_id,omitempty"`
	Class            string `json:"class"`
	Owner            string `json:"owner"`
	Mode             string `json:"mode"`
	Summary          string `json:"summary"`
	SuggestedCommand string `json:"suggested_command,omitempty"`
}
