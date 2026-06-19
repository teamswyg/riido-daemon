package main

type options struct {
	Repo        string
	Manifest    string
	WriteDoc    bool
	CheckDoc    bool
	EvidenceOut string
}

type manifest struct {
	SchemaVersion    string   `json:"schema_version"`
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	GeneratedDoc     string   `json:"generated_doc"`
	Workflow         string   `json:"workflow"`
	EvidenceArtifact string   `json:"evidence_artifact"`
	Summary          []string `json:"summary"`
	Pages            []page   `json:"pages"`
	Assertions       []string `json:"assertions"`
}

type page struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	GeneratedDoc string     `json:"generated_doc"`
	ProviderID   string     `json:"provider_id,omitempty"`
	Summary      []string   `json:"summary"`
	Facts        []string   `json:"facts"`
	TableColumns []string   `json:"table_columns,omitempty"`
	TableRows    [][]string `json:"table_rows,omitempty"`
	Artifacts    []string   `json:"artifacts"`
	Links        []link     `json:"links,omitempty"`
}
