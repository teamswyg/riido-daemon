package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion    string            `json:"schema_version"`
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	GeneratedDoc     string            `json:"generated_doc"`
	Workflow         string            `json:"workflow"`
	EvidenceArtifact string            `json:"evidence_artifact"`
	Loop             evidenceLoop      `json:"loop"`
	Summary          []string          `json:"summary"`
	DetailPages      []detailLink      `json:"detail_pages"`
	RelatedPages     []detailLink      `json:"related_pages"`
	Fragments        map[string]string `json:"fragments"`
	SourceChecks     []sourceCheck     `json:"source_checks"`
	Assertions       []string          `json:"assertions"`
	Details          []detailDoc       `json:"-"`
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
