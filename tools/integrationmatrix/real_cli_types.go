package main

type realCLIObservation struct {
	SchemaVersion    string         `json:"schema_version"`
	ID               string         `json:"id"`
	Title            string         `json:"title"`
	GeneratedDoc     string         `json:"generated_doc"`
	Workflow         string         `json:"workflow"`
	EvidenceArtifact string         `json:"evidence_artifact"`
	Providers        []realProvider `json:"providers"`
}

type realProvider struct {
	ID                string `json:"id"`
	DisplayName       string `json:"display_name"`
	DefaultExecutable string `json:"default_executable"`
	OverrideEnv       string `json:"override_env"`
	GoPackage         string `json:"go_package"`
	TestRegex         string `json:"test_regex"`
}
