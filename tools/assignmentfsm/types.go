package main

type Manifest struct {
	SchemaVersion      string           `json:"schema_version"`
	ID                 string           `json:"id"`
	Title              string           `json:"title"`
	LoopSource         string           `json:"loop_source,omitempty"`
	GeneratedDoc       string           `json:"generated_doc"`
	Workflow           string           `json:"workflow"`
	EvidenceArtifact   string           `json:"evidence_artifact"`
	SourcePackage      string           `json:"source_package"`
	EvidenceManifest   string           `json:"evidence_manifest"`
	Purpose            string           `json:"purpose"`
	SourceChecks       []SourceCheck    `json:"source_checks"`
	ForbiddenDocTokens []string         `json:"forbidden_doc_tokens"`
	RelatedSections    []RelatedSection `json:"related_sections"`
}

type SourceCheck struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Contains string `json:"contains"`
}

type RelatedSection struct {
	Title string `json:"title"`
	Path  string `json:"path"`
}
