package main

type evidence struct {
	SchemaVersion string              `json:"schema_version"`
	ManifestID    string              `json:"manifest_id"`
	GeneratedDoc  string              `json:"generated_doc"`
	Workflow      string              `json:"workflow"`
	SourceChecks  []sourceCheckResult `json:"source_checks"`
	Problems      []string            `json:"problems"`
}

type sourceCheckResult struct {
	Name string `json:"name"`
	File string `json:"file"`
	Pass bool   `json:"pass"`
}
