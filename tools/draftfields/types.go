package main

type Manifest struct {
	SchemaVersion    string           `json:"schema_version"`
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	AllowedDoc       string           `json:"allowed_doc"`
	ForbiddenDoc     string           `json:"forbidden_doc"`
	Workflow         string           `json:"workflow"`
	EvidenceArtifact string           `json:"evidence_artifact"`
	ForbiddenScope   []string         `json:"forbidden_scope"`
	AllowedFields    []AllowedField   `json:"allowed_fields"`
	ForbiddenFields  []ForbiddenField `json:"forbidden_fields"`
	Assertions       []string         `json:"assertions"`
}

type AllowedField struct {
	Field    string `json:"field"`
	Status   string `json:"status"`
	Meaning  string `json:"meaning"`
	Source   string `json:"source,omitempty"`
	Contains string `json:"contains,omitempty"`
}

type ForbiddenField struct {
	Field    string   `json:"field"`
	FilledBy string   `json:"filled_by"`
	Reason   string   `json:"reason"`
	Tokens   []string `json:"tokens"`
}
