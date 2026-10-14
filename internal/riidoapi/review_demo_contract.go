package riidoapi

type ReviewDemoRequest struct {
	DistributionChannel      string `json:"distribution_channel"`
	ReviewDemoConsentGranted bool   `json:"review_demo_consent_granted"`
}

type ReviewDemoResponse struct {
	SchemaVersion            string   `json:"schema_version"`
	DistributionChannel      string   `json:"distribution_channel"`
	Enabled                  bool     `json:"enabled"`
	Surfaces                 []string `json:"surfaces"`
	ProviderStatusMode       string   `json:"provider_status_mode"`
	ProviderExecutionAllowed bool     `json:"provider_execution_allowed"`
	TelemetrySyncAllowed     bool     `json:"telemetry_sync_allowed"`
	LocalOnly                bool     `json:"local_only"`
}
