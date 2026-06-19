package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion    string        `json:"schema_version"`
	ID               string        `json:"id"`
	Title            string        `json:"title"`
	Workflow         string        `json:"workflow"`
	EvidenceArtifact string        `json:"evidence_artifact"`
	Fragments        []string      `json:"fragments"`
	SourceChecks     []sourceCheck `json:"source_checks"`
	Assertions       []string      `json:"assertions"`
	Pages            []page        `json:"-"`
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
