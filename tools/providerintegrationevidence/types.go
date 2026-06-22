package main

type manifest struct {
	SchemaVersion    string     `json:"schema_version"`
	LoopSource       string     `json:"loop_source,omitempty"`
	ID               string     `json:"id"`
	Title            string     `json:"title"`
	GeneratedDoc     string     `json:"generated_doc"`
	Workflow         string     `json:"workflow"`
	EvidenceArtifact string     `json:"evidence_artifact"`
	Providers        []provider `json:"providers"`
}

type provider struct {
	ID                string `json:"id"`
	DisplayName       string `json:"display_name"`
	DefaultExecutable string `json:"default_executable"`
	OverrideEnv       string `json:"override_env"`
	GoPackage         string `json:"go_package"`
	TestRegex         string `json:"test_regex"`
}

type evidenceFile struct {
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
	ID                 string         `json:"id"`
	Available          bool           `json:"available"`
	ExecutableRef      string         `json:"executable_ref"`
	ExecutablePath     string         `json:"executable_path,omitempty"`
	Version            string         `json:"version,omitempty"`
	IntegrationStatus  string         `json:"integration_status"`
	IntegrationCommand string         `json:"integration_command"`
	FailureSummary     string         `json:"failure_summary,omitempty"`
	Observed           map[string]any `json:"observed,omitempty"`
	Repair             *repair        `json:"repair,omitempty"`
}

type repair struct {
	Class            string `json:"class"`
	Owner            string `json:"owner"`
	Mode             string `json:"mode"`
	Summary          string `json:"summary"`
	SuggestedCommand string `json:"suggested_command,omitempty"`
}
