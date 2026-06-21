package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion       string            `json:"schema_version"`
	ID                  string            `json:"id"`
	Title               string            `json:"title"`
	LoopSource          string            `json:"loop_source,omitempty"`
	GeneratedDoc        string            `json:"generated_doc"`
	Workflow            string            `json:"workflow"`
	EvidenceArtifact    string            `json:"evidence_artifact"`
	Summary             []string          `json:"summary"`
	Responsibilities    []string          `json:"responsibilities"`
	NonResponsibilities []string          `json:"non_responsibilities"`
	Invariants          []string          `json:"invariants"`
	DetailPages         []detailLink      `json:"detail_pages"`
	Fragments           map[string]string `json:"fragments"`
	SourceChecks        []sourceCheck     `json:"source_checks"`
	Assertions          []string          `json:"assertions"`
	Details             []detailDoc       `json:"-"`
}

type detailLink struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

type sourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}
