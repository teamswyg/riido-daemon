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
	SchemaVersion string             `json:"schema_version"`
	ID            string             `json:"id"`
	ObservedAt    string             `json:"observed_at"`
	Status        string             `json:"status"`
	Providers     []providerEvidence `json:"providers"`
}

type providerEvidence struct {
	ID                 string `json:"id"`
	Available          bool   `json:"available"`
	ExecutableRef      string `json:"executable_ref"`
	Version            string `json:"version,omitempty"`
	IntegrationStatus  string `json:"integration_status"`
	IntegrationCommand string `json:"integration_command"`
	FailureSummary     string `json:"failure_summary,omitempty"`
}
