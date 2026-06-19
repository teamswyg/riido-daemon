package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion        string        `json:"schema_version"`
	ID                   string        `json:"id"`
	Title                string        `json:"title"`
	GeneratedDoc         string        `json:"generated_doc"`
	Workflow             string        `json:"workflow"`
	EvidenceArtifact     string        `json:"evidence_artifact"`
	Summary              []string      `json:"summary"`
	CompatibilityMarkers []string      `json:"compatibility_markers"`
	Parts                []link        `json:"parts"`
	RelatedPages         []link        `json:"related_pages"`
	SourceChecks         []sourceCheck `json:"source_checks"`
	Assertions           []string      `json:"assertions"`
}

type link struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}

type sourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
